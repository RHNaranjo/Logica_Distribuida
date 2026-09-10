package mundos

import (
	"net/http"

	"logica/internal/web"
)

const nombreModulo = "mundos"

// Controlador agrupa las rutas del módulo.
// Todavía no tiene campos; en la Fase 5 recibirá el repositorio del historial.
type Controlador struct{}

func NuevoControlador() *Controlador {
	return &Controlador{}
}

func (c *Controlador) Rutas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", c.Salud)
	mux.HandleFunc("POST /resolver", c.Resolver)

	return mux
}

func (c *Controlador) Salud(w http.ResponseWriter, r *http.Request) {
	web.JSON(w, http.StatusOK, map[string]any{
		"modulo": nombreModulo,
		// ConNodo ya escribió este header antes de llegar aquí,
		// así que la instancia se entera de su propio nombre sin
		// que haya que pasárselo por el constructor.
		"instancia": w.Header().Get("X-Nodo"),
		"ok":        true,
	})
}

func (c *Controlador) Resolver(w http.ResponseWriter, r *http.Request) {
	web.Fallo(w, http.StatusNotImplemented, "el módulo de mundos posibles se implementará después")
}
