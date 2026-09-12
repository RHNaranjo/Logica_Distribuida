package puerta

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"logica/internal/web"
)

// Por aquí entra y sale todo
type Controlador struct {
	proxy *httputil.ReverseProxy
}

// El proxy se arma una sola vez cuando se arranca
func NuevoControlador(destino *url.URL) *Controlador {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(destino)
			pr.Out.URL.Path = rutaEnLoadBalancer(pr.In.URL.Path)
			pr.Out.URL.RawPath = ""
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[ERROR] El loadbalancer no respondió: %v", err)
			web.Fallo(w, http.StatusBadGateway, "el loadbalancer no responde")
		},
	}

	return &Controlador{proxy: proxy}
}

func (c *Controlador) Rutas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", c.Salud)
	mux.HandleFunc("GET /api/estado", c.Reenviar)
	mux.HandleFunc("/api/{modulo}/{ruta...}", c.Reenviar)

	return mux
}

// Obtener la salud desde el middleware
func (c *Controlador) Salud(w http.ResponseWriter, r *http.Request) {
	web.JSON(w, http.StatusOK, map[string]any{
		"servicio": "middleware",
		"ok":       true,
	})
}

// Reenviar el contenido con el proxy
func (c *Controlador) Reenviar(w http.ResponseWriter, r *http.Request) {
	c.proxy.ServeHTTP(w, r)
}

// Traducir la ruta pública a la interna
// /api/estado -> /instancias
// /api/tablas/salud -> /tablas/salud
func rutaEnLoadBalancer(publica string) string {
	if publica == "/api/estado" {
		return "/instancias"
	}

	return strings.TrimPrefix(publica, "/api")
}
