package main

import (
	"log"
	"net/http"

	"logica/internal/config"
	"logica/internal/middleware"
)

func main() {
	directorio := config.DirectorioWeb("./frontend/dist")

	mux := http.NewServeMux()

	// Servir los archivos de la página
	mux.Handle("GET /", http.FileServer(http.Dir(directorio)))

	var handler http.Handler = mux
	handler = middleware.ConNodo("frontend", handler)
	handler = middleware.ConLog("frontend", handler)

	puerto := config.PuertoHTTP("3000")
	log.Printf("[INFO] Frontend sirviendo %q en %s", directorio, puerto)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El frontend se detuvo: %v", err)
	}
}
