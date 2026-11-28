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

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/filters"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"

	"logica/internal/rpc"
	"logica/internal/tlsx"
)

type Opciones struct {
	Modulo     string // mundos, tablas o deducción natural
	Nombre     string // nombre de la instancia de k8s
	Instancias int
	PuertoBase int
}

// Levantar N servidores gRPC (uno por gorutina)
func Arrancar(op Opciones, nuevo func(instancia string) rpc.ModuloServer) error {
	errores := make(chan error, op.Instancias)

	// Leer el certificado
	creds, err := tlsx.Servidor()
	if err != nil {
		return fmt.Errorf("no se pudo cargar el mTLS: %w", err)
	}

	for i := 0; i < op.Instancias; i++ {
		puerto := op.PuertoBase + i

		// 1 vs muchas instancias
		nombre := op.Nombre
		if op.Instancias > 1 {
			nombre = fmt.Sprintf("%s-%d", op.Nombre, i)
		}

		// Abrir el puerto de esta instancia
		escucha, err := net.Listen("tcp", fmt.Sprintf(":%d", puerto))
		if err != nil {
			return fmt.Errorf("no se pudo abrir el puerto %d: %w", puerto, err)
		}

		// Crear el servidor con el interceptor de cada llamada
		// Cerrar cada conexión a los 30 s para que el cliente pregunte otra vez al DNS
		srv := grpc.NewServer(
			grpc.Creds(creds),
			// Crear un span por llamada, menos en los health checks
			grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithFilter(filters.Not(filters.HealthCheck())))),
			grpc.UnaryInterceptor(interceptor(nombre)),
			grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: 30 * time.Second}),
		)

		// Registrar el módulo, la salud y la reflexión para grpcurl
		rpc.RegisterModuloServer(srv, nuevo(nombre))
		healthpb.RegisterHealthServer(srv, health.NewServer())
		reflection.Register(srv)

		go func() {
			log.Printf("[INFO] %s escuchando gRPC en :%d", nombre, puerto)
			errores <- fmt.Errorf("la instancia %s se detuvo: %w", nombre, srv.Serve(escucha))
		}()
	}

	// Escuchar si se apaga el sistema
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

		// Leer quién llamó según el nombre del certificado
		cliente := "-"
		if p, ok := peer.FromContext(ctx); ok {
			if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok && len(tlsInfo.State.PeerCertificates) > 0 {
				cliente = tlsInfo.State.PeerCertificates[0].Subject.CommonName
			}
		}

		resp, err := handler(ctx, req)

		// EJEMPLO: [tablas-0] p8 /logica.Modulo/Resolver de middleware (1.2ms)
		log.Printf("[%s] %s %s de %s (%s)", nombre, peticion, info.FullMethod, cliente, time.Since(inicio))

		return resp, err
	}
}
