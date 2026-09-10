package tablas

import (
	"net/http"

	"logica/internal/web"
)

const nombreModulo = "tablas"

// Aguprar las rutas del módulo
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
		// La instancia se entera por su cuenta de su nombre
		"instancia": w.Header().Get("X-Nodo"),
		"ok":        true,
	})
}

func (c *Controlador) Resolver(w http.ResponseWriter, r *http.Request) {
	web.Fallo(w, http.StatusNotImplemented, "tablas de verdad viene después :)")
}
