package main

import (
	"log"
	"net/http"
	"net/url"

	"logica/internal/config"
	"logica/internal/middleware"
	"logica/internal/puerta"
)

func main() {
	url_lb := config.URLLoadBalancer("http://localhost:8081")

	// Espera un loadbalancer válido
	destino, err := url.Parse(url_lb)
	if err != nil || destino.Scheme == "" || destino.Host == "" {
		log.Fatalf("[ERROR] LOADBALANCER_URL inválida: %q", url_lb)
	}

	ctrl := puerta.NuevoControlador(destino)

	// CORS(Peticion(Log(Nodo(rutas))))
	var handler http.Handler = ctrl.Rutas()
	handler = middleware.ConNodo("middleware", handler) // Sólo marca la respuesta
	handler = middleware.ConLog("middleware", handler)
	handler = middleware.ConPeticion(handler) // Después de ConLog para que el ID ya exista
	handler = middleware.ConCORS(handler)     // Atrapa el OPTIONS

	puerto := config.PuertoHTTP("8080")
	log.Printf("[INFO] middleware escuchando en %s. Se reenvía a %s", puerto, destino)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El middleware se detuvo: %v", err)
	}
}
