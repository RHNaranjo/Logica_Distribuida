package main

import (
	"log"
	"net/http"

	"mundos_posibles/internal/config"
	"mundos_posibles/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	// El /{$} hace match con la raiz, para que no atrape cualquier cosa
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if _, err := w.Write([]byte(paginaProvisional)); err != nil {
			log.Printf("[ERROR] No se pudo escribir la respuesta: %v", err)
		}
	})

	handler := middleware.ConLog("coordinador", middleware.ConNodo("coordinador", mux))

	puerto := config.PuertoHTTP()
	log.Printf("[INFO] Coordinador escuchando en %s", puerto)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El coordinador se detuvo: %v", err)
	}
}

// Pagina provisional en lo que se hace el resto
const paginaProvisional = `<!doctype html>
<html lang="es">
<head><meta charset="utf-8"><title>Mundos Posibles</title></head>
<body>
<h1>Coordinador</h1>
<p>Esqueleto de la Fase 1. Todavía no hay modelo de Kripke que calcular.</p>
</body>
</html>`
