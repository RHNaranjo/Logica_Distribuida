package puerta

import (
	"cmp"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"logica/internal/rpc"
	"logica/internal/web"
)

const (
	tiempoLimite = 10 * time.Second
	cuerpoMaximo = 1 << 20 // un MB
)

// Lo que el frontend ve de cada instancia
type Instancia struct {
	Instancia string `json:"instancia"`
	Sana      bool   `json:"sana"`
	Atendidas int    `json:"atendidas"`
}

// Traducir el HTTP del navegador a gRPC
type Controlador struct {
	modulo rpc.ModuloClient
	salud  healthpb.HealthClient

	mu        sync.Mutex
	atendidas map[string]map[string]int // "tablas" -> "tablas-1" -> 12
}

// Los dos clientes comparten la misma conexión con nginx
func NuevoControlador(conexion *grpc.ClientConn) *Controlador {
	return &Controlador{
		modulo:    rpc.NewModuloClient(conexion),
		salud:     healthpb.NewHealthClient(conexion),
		atendidas: make(map[string]map[string]int),
	}
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

	ctx, cancel := llamada(r, modulo)
	defer cancel()

	var cabeceras metadata.MD
	resp, err := c.salud.Check(ctx, &healthpb.HealthCheckRequest{}, grpc.Header(&cabeceras))
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

	ctx, cancel := llamada(r, modulo)
	defer cancel()

	var cabeceras metadata.MD
	salida, err := c.modulo.Resolver(ctx, entrada, grpc.Header(&cabeceras))
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

	ctx, cancel := llamada(r, modulo)
	defer cancel()

	var cabeceras metadata.MD
	lista, err := c.modulo.Historial(ctx, &rpc.PeticionHistorial{Limite: int32(limite)}, grpc.Header(&cabeceras))
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

// Devolver cuántas peticiones ha atendido cada instancia
func (c *Controlador) Estado(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	estado := make(map[string][]Instancia)

	for modulo, cuentas := range c.atendidas {
		for nombre, n := range cuentas {
			estado[modulo] = append(estado[modulo], Instancia{Instancia: nombre, Sana: true, Atendidas: n})
		}

		// Ordenar por nombre para que no bailen en la página
		slices.SortFunc(estado[modulo], func(a, b Instancia) int {
			return cmp.Compare(a.Instancia, b.Instancia)
		})
	}

	web.JSON(w, http.StatusOK, estado)
}

// Copiar el recorrido y la instancia que respondió, y contarla
func (c *Controlador) marcar(w http.ResponseWriter, modulo string, cabeceras metadata.MD) string {
	// Agregar los nodos que se cruzaron (nginx-lb)
	for _, nodo := range cabeceras.Get("x-nodo") {
		w.Header().Add("X-Nodo", nodo)
	}

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

// Preparar la llamada: tiempo límite y headers para nginx y la instancia
func llamada(r *http.Request, modulo string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), tiempoLimite)

	// nginx elige el grupo de instancias con "modulo"; la instancia anota "x-peticion"
	ctx = metadata.AppendToOutgoingContext(ctx, "modulo", modulo, "x-peticion", r.Header.Get("X-Peticion"))

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
