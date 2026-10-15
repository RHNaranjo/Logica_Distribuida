package servidor

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"logica/internal/consul"
	"logica/internal/rpc"
)

type Opciones struct {
	Modulo     string // mundos, tablas o deducción natural
	Instancias int
	PuertoBase int
	Host       string
	Consul     string // host:puerto
}

// Levantar N servidores gRPC, uno por gorutina. Se bloquea hasta que alguno de ellos se caiga
func Arrancar(op Opciones, nuevo func(instancia string) rpc.ModuloServer) error {
	errores := make(chan error, op.Instancias)

	var nombres []string

	// Borrar las instancias si se sale por error o por apagar
	defer func() {
		for _, nombre := range nombres {
			consul.Borrar(op.Consul, nombre)
		}
	}()

	for i := 0; i < op.Instancias; i++ {
		puerto := op.PuertoBase + i
		nombre := fmt.Sprintf("%s-%d", op.Modulo, i)

		// Abrir el puerto de esta instancia
		escucha, err := net.Listen("tcp", fmt.Sprintf(":%d", puerto))
		if err != nil {
			return fmt.Errorf("no se pudo abrir el puerto %d: %w", puerto, err)
		}

		// Crear el servidor con el interceptor de cada llamada
		srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor(nombre)))

		// Registrar el modulo, la salud y la reflexión para grpcurl
		rpc.RegisterModuloServer(srv, nuevo(nombre))
		healthpb.RegisterHealthServer(srv, health.NewServer())
		reflection.Register(srv)

		go func() {
			log.Printf("[INFO] %s escuchando gRPC en :%d", nombre, puerto)
			errores <- fmt.Errorf("la instancia %s se detuvo: %w", nombre, srv.Serve(escucha))
		}()

		// Darse de alta en Consul
		nombres = append(nombres, nombre)
		go consul.Registrar(op.Consul, consul.Servicio{ID: nombre, Nombre: op.Modulo, Host: op.Host, Puerto: puerto})
	}

	// Escuchar el apagado del docker stop
	apagado := make(chan os.Signal, 1)
	signal.Notify(apagado, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errores:
		return err
	case <-apagado:
		log.Printf("[INFO] Apagando %s", op.Modulo)
		return nil
	}
}

func interceptor(nombre string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		inicio := time.Now()

		// Avisar quién respondió; llega al middleware como un header
		grpc.SetHeader(ctx, metadata.Pairs("x-instancia", nombre))

		// Leer el id de petición que manda el middleware
		peticion := "-"
		if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get("x-peticion")) > 0 {
			peticion = md.Get("x-peticion")[0]
		}

		resp, err := handler(ctx, req)

		// EJEMPLO: [tablas-0] p8 /logica.Modulo/Resolver (1.2ms)
		log.Printf("[%s] %s %s (%s)", nombre, peticion, info.FullMethod, time.Since(inicio))

		return resp, err
	}
}
