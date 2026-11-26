package main

import (
	"log"
	"net/http"

	"logica/internal/config"
	"logica/internal/middleware"
	"logica/internal/puerta"
)

func main() {
	// El DNs de Kube dice lo que hay en cada módulo
	destinos := map[string]string{
		"mundos":    "mundos:9001",
		"tablas":    "tablas:9011",
		"deduccion": "deduccion:9021",
	}

	ctrl, err := puerta.NuevoControlador(destinos)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	// Peticion(Log(Nodo(rutas)))
	var handler http.Handler = ctrl.Rutas()
	handler = middleware.ConNodo("middleware", handler)
	handler = middleware.ConLog("middleware", handler)
	handler = middleware.ConPeticion(handler) // Después de ConLog para que el ID exista

	puerto := config.PuertoHTTP("8080")
	log.Printf("[INFO] middleware escuchando en %s. Módulos: %v", puerto, destinos)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El middleware se detuvo: %v", err)
	}
}
