package puerta

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"logica/internal/rpc"
	"logica/internal/tlsx"
	"logica/internal/web"
)

const (
	tiempoLimite = 10 * time.Second
	cuerpoMaximo = 1 << 20 // un MB
)

// Repartir las llamadas entre todas las IPs
const balanceo = `{"loadBalancingConfig":[{"round_robin": {}}]}`

// Lo que el frontend ve de cada instancia
type Instancia struct {
	Instancia string `json:"instancia"`
	Sana      bool   `json:"sana"`
	Atendidas int    `json:"atendidas"`
}

// Traducir el HTTP a gRPC
type Controlador struct {
	destinos   map[string]string                           // "tablas" -> "tablas:9011"
	conexiones map[string]*grpc.ClientConn                 // una por módulo
	creds      map[string]credentials.TransportCredentials // "tablas" -> mTLS con ServerName "tablas"
	mu         sync.Mutex
	atendidas  map[string]map[string]int //"tablas" -> "tablas-7cssas8asc-weerq8" -> 12
}

// Abrir una conexión gRPC por módulo, balanceada del lado del cliente
func NuevoControlador(destinos map[string]string) (*Controlador, error) {
	c := &Controlador{
		destinos:   destinos,
		conexiones: make(map[string]*grpc.ClientConn),
		creds:      make(map[string]credentials.TransportCredentials),
		atendidas:  make(map[string]map[string]int),
	}

	for modulo, destino := range destinos {
		// El nombre del Service debe estar en el SAN del certificado del módulo
		host, _, err := net.SplitHostPort(destino)
		if err != nil {
			return nil, fmt.Errorf("destino inválido para %s (%q): %w", modulo, destino, err)
		}

		creds, err := tlsx.Cliente(host)
		if err != nil {
			return nil, fmt.Errorf("no se pudo cargar el mTLS: %w", err)
		}

		// dns:/// pide todas las IPS
		conexion, err := grpc.NewClient(
			"dns:///"+destino,
			grpc.WithTransportCredentials(creds),
			grpc.WithDefaultServiceConfig(balanceo),
		)
		if err != nil {
			return nil, fmt.Errorf("destino inválido para %s (%q): %w", modulo, destino, err)
		}

		c.conexiones[modulo] = conexion
		c.creds[modulo] = creds
	}

	return c, nil
}

func (c *Controlador) Rutas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", c.Salud)
	mux.HandleFunc("GET /api/estado", c.Estado)
	mux.HandleFunc("GET /api/{modulo}/salud", c.SaludModulo)
	mux.HandleFunc("POST /api/{modulo}/resolver", c.Resolver)
	mux.HandleFunc("GET /api/{modulo}/historial", c.Historial)

	return mux
}

// Confirmar que el middleware está vivo
func (c *Controlador) Salud(w http.ResponseWriter, r *http.Request) {
	web.JSON(w, http.StatusOK, map[string]any{
		"servicio": "middleware",
		"ok":       true,
	})
}

// Preguntar a una instancia del módulo si está sana
func (c *Controlador) SaludModulo(w http.ResponseWriter, r *http.Request) {
	modulo := r.PathValue("modulo")

	conexion, ok := c.conexion(w, modulo)
	if !ok {
		return
	}

	ctx, cancel := llamada(r)
	defer cancel()

	var cabeceras metadata.MD
	resp, err := healthpb.NewHealthClient(conexion).Check(ctx, &healthpb.HealthCheckRequest{}, grpc.Header(&cabeceras))
	instancia := c.marcar(w, modulo, cabeceras)

	if err != nil {
		fallo(w, err)
		return
	}

	web.JSON(w, http.StatusOK, map[string]any{
		"modulo":    modulo,
		"instancia": instancia,
		"estado":    resp.GetStatus().String(),
	})
}

// Mandar la entrada del navegador al módulo
func (c *Controlador) Resolver(w http.ResponseWriter, r *http.Request) {
	modulo := r.PathValue("modulo")

	conexion, ok := c.conexion(w, modulo)
	if !ok {
		return
	}

	// Leer el cuerpo con tope de memoria
	cuerpo, err := io.ReadAll(http.MaxBytesReader(w, r.Body, cuerpoMaximo))
	if err != nil {
		web.Fallo(w, http.StatusBadRequest, "no se pudo leer el cuerpo de la petición")
		return
	}

	// Convertir el JSON a Struct
	entrada := &structpb.Struct{}
	if err := protojson.Unmarshal(cuerpo, entrada); err != nil {
		web.Fallo(w, http.StatusBadRequest, "el cuerpo debe ser un JSON")
		return
	}

	ctx, cancel := llamada(r)
	defer cancel()

	var cabeceras metadata.MD
	salida, err := rpc.NewModuloClient(conexion).Resolver(ctx, entrada, grpc.Header(&cabeceras))
	instancia := c.marcar(w, modulo, cabeceras)

	if err != nil {
		fallo(w, err)
		return
	}

	// Regresar el Struct como JSON normal
	crudo, err := protojson.Marshal(salida)
	if err != nil {
		web.Fallo(w, http.StatusInternalServerError, "no se pudo leer la respuesta del módulo")
		return
	}

	web.JSON(w, http.StatusOK, map[string]any{
		"instancia": instancia,
		"resultado": json.RawMessage(crudo),
	})
}

// Pedir los últimos registros del módulo
func (c *Controlador) Historial(w http.ResponseWriter, r *http.Request) {
	modulo := r.PathValue("modulo")

	conexion, ok := c.conexion(w, modulo)
	if !ok {
		return
	}

	// Leer el límite si viene; 0 significa el que quiera
	var limite int64
	if crudo := r.URL.Query().Get("limite"); crudo != "" {
		n, err := strconv.ParseInt(crudo, 10, 32)
		if err != nil {
			web.Fallo(w, http.StatusBadRequest, "el límite debe ser un número entero")
			return
		}

		limite = n
	}

	ctx, cancel := llamada(r)
	defer cancel()

	var cabeceras metadata.MD
	lista, err := rpc.NewModuloClient(conexion).Historial(ctx, &rpc.PeticionHistorial{Limite: int32(limite)}, grpc.Header(&cabeceras))
	c.marcar(w, modulo, cabeceras)

	if err != nil {
		fallo(w, err)
		return
	}

	crudo, err := protojson.Marshal(lista)
	if err != nil {
		web.Fallo(w, http.StatusInternalServerError, "no se pudo leer el historial del módulo")
		return
	}

	web.JSON(w, http.StatusOK, json.RawMessage(crudo))
}

// Preguntar a cada pod de cada módulo si está sano
func (c *Controlador) Estado(w http.ResponseWriter, r *http.Request) {
	estado := make(map[string][]Instancia)

	for modulo, destino := range c.destinos {
		// Sin pods, el módulo aparece con la lista vacía
		estado[modulo] = []Instancia{}

		host, puerto, err := net.SplitHostPort(destino)
		if err != nil {
			continue
		}

		// El Service headless responde con la IP de cada pod listo
		ips, err := net.DefaultResolver.LookupHost(r.Context(), host)
		if err != nil {
			log.Printf("[AVISO] No se pudo resolver %s: %v", host, err)
			continue
		}

		for _, ip := range ips {
			estado[modulo] = append(estado[modulo], c.revisar(r.Context(), modulo, net.JoinHostPort(ip, puerto)))
		}

		// Ordenar por nombre
		slices.SortFunc(estado[modulo], func(a, b Instancia) int {
			return cmp.Compare(a.Instancia, b.Instancia)
		})
	}

	web.JSON(w, http.StatusOK, estado)
}

// Hacer el health check a un solo pod y leer su nombre
func (c *Controlador) revisar(ctx context.Context, modulo, direccion string) Instancia {
	// Si no contesta, por lo menos se ve su IP
	inst := Instancia{Instancia: direccion}

	// Conexión directa a esa IP, sin balanceo, pero con el ServerName del módulo
	conexion, err := grpc.NewClient(direccion, grpc.WithTransportCredentials(c.creds[modulo]))
	if err != nil {
		return inst
	}
	defer conexion.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var cabeceras metadata.MD
	resp, err := healthpb.NewHealthClient(conexion).Check(ctx, &healthpb.HealthCheckRequest{}, grpc.Header(&cabeceras))

	if nombres := cabeceras.Get("x-instancia"); len(nombres) > 0 {
		inst.Instancia = nombres[0]
	}
	inst.Sana = err == nil && resp.GetStatus() == healthpb.HealthCheckResponse_SERVING

	c.mu.Lock()
	defer c.mu.Unlock()
	inst.Atendidas = c.atendidas[modulo][inst.Instancia]

	return inst
}

// Buscar la conexión del módulo; responder 404 si no existe
func (c *Controlador) conexion(w http.ResponseWriter, modulo string) (*grpc.ClientConn, bool) {
	conexion, ok := c.conexiones[modulo]
	if !ok {
		web.Fallo(w, http.StatusNotFound, "no existe el módulo "+modulo)
	}

	return conexion, ok
}

// Copiar la instancia que respondió al recorrido, y contarla
func (c *Controlador) marcar(w http.ResponseWriter, modulo string, cabeceras metadata.MD) string {
	nombres := cabeceras.Get("x-instancia")
	if len(nombres) == 0 {
		return ""
	}

	instancia := nombres[0]
	w.Header().Set("X-Instancia", instancia)
	w.Header().Add("X-Nodo", instancia)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.atendidas[modulo] == nil {
		c.atendidas[modulo] = make(map[string]int)
	}
	c.atendidas[modulo][instancia]++

	return instancia
}

// Preparar la llamada: tiempo límite y el id de petición para la instancia
func llamada(r *http.Request) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), tiempoLimite)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-peticion", r.Header.Get("X-Peticion"))

	return ctx, cancel
}

// Traducir el error gRPC a un código HTTP
func fallo(w http.ResponseWriter, err error) {
	codigo := http.StatusBadGateway

	switch status.Code(err) {
	case codes.InvalidArgument:
		codigo = http.StatusBadRequest
	case codes.Unimplemented:
		codigo = http.StatusNotImplemented
	case codes.Unavailable:
		codigo = http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		codigo = http.StatusGatewayTimeout
	case codes.Internal:
		codigo = http.StatusInternalServerError
	}

	web.Fallo(w, codigo, status.Convert(err).Message())
}
