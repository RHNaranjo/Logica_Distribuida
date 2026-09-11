package loadbalancer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"logica/internal/servidor"
	"logica/internal/web"
)

type Controlador struct {
	reg *Registro
}

func NuevoControlador(reg *Registro) *Controlador {
	return &Controlador{reg: reg}
}

func (c *Controlador) Rutas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", c.Salud)
	mux.HandleFunc("POST /registrar", c.Registrar)
	mux.HandleFunc("GET /instancias", c.Instancias)

	// El proxy reenvía GET, POST o lo que pueda llegar
	mux.HandleFunc("/{modulo}/{ruta...}", c.Reenviar)

	return mux
}

// Confirmar que el loadbalancer está vivo
func (c *Controlador) Salud(w http.ResponseWriter, r *http.Request) {
	web.JSON(w, http.StatusOK, map[string]any{
		"servicio": "loadbalancer",
		"ok":       true,
	})
}

// Decodificar el Anuncio, validar que traiga todo bien hecho y lo guarda
func (c *Controlador) Registrar(w http.ResponseWriter, r *http.Request) {
	var a servidor.Anuncio

	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		web.Fallo(w, http.StatusBadRequest, "el cuerpo del texto no es un JSON válido")
		return
	}

	u, err := url.Parse(a.URL)
	if a.Modulo == "" || a.Instancia == "" || err != nil || u.Scheme == "" || u.Host == "" {
		web.Fallo(w, http.StatusBadRequest, "hace falta información (módulo, instancia o URL válida)")
		return
	}

	c.reg.Agregar(a.Modulo, a.Instancia, a.URL)
	w.WriteHeader(http.StatusNoContent)
}

// Devolver una copia del directorio en frmato JSON (quién está registrado vs sano)
func (c *Controlador) Instancias(w http.ResponseWriter, r *http.Request) {
	web.JSON(w, http.StatusOK, c.reg.Foto())
}

// Elige una instancia sana y le pasa la petición
func (c *Controlador) Reenviar(w http.ResponseWriter, r *http.Request) {
	modulo := r.PathValue("modulo")
	ruta := r.PathValue("ruta")

	inst, ok := c.reg.Elegir(modulo)
	if !ok {
		web.Fallo(w, http.StatusServiceUnavailable, fmt.Sprintf("No hay instancias sanas en el módulo %q", modulo))
		return
	}

	destino, err := url.Parse(inst.URL)
	if err != nil {
		web.Fallo(w, http.StatusInternalServerError, "La instancia elegida tiene una URL inválida")
		return
	}

	proxy := &httputil.ReverseProxy{
		// Armar la petición de salida a partir de la entrada
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(destino)
			pr.Out.URL.Path = "/" + ruta
			pr.Out.URL.RawPath = ""
		},

		// Si la instancia no contesta, se saca de la rotacoión
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[ERROR] %s no contestó: %v", inst.Nombre, err)
			c.reg.MarcarSalud(modulo, inst.Nombre, false)
			web.Fallo(w, http.StatusBadGateway, fmt.Sprintf("la instancia %s no respondió", inst.Nombre))
		},
	}

	w.Header().Set("X-Instancia", inst.Nombre)
	proxy.ServeHTTP(w, r)
}
