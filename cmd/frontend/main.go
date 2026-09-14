package main

import (
	"log"
	"net/http"

	"logica/internal/config"
	"logica/internal/middleware"
	"logica/internal/web"
)

func main() {
	directorio := config.DirectorioWeb("./frontend/dist")
	publico := config.MiddlewarePublico("http://localhost:8080")

	mux := http.NewServeMux()

	// El navegador no puede leer env vars, el archivo entrega la dirección sin recompilar
	mux.HandleFunc("GET /config.json", func(w http.ResponseWriter, r *http.Request) {
		web.JSON(w, http.StatusOK, map[string]string{"middleware": publico})
	})

	// De aquí en adelante hay puro archivo estático
	mux.Handle("GET /", http.FileServer(http.Dir(directorio)))

	var handler http.Handler = mux
	handler = middleware.ConNodo("frontend", handler)
	handler = middleware.ConLog("frontend", handler)

	puerto := config.PuertoHTTP("3000")
	log.Printf("[INFO] Frontend sirviendo %q en %s (middleware público: %s)", directorio, puerto, publico)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El frontend se detuvo: %v", err)
	}
}
