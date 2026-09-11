package main

import (
	"log"
	"net/http"
	"time"

	"logica/internal/config"
	"logica/internal/loadbalancer"
	"logica/internal/middleware"
)

func main() {
	reg := loadbalancer.NuevoRegistro()

	// El vigilante arranca ANTES de ListenAndServe: esa llamada bloquea
	// para siempre y nada de lo que esté debajo llegaría a ejecutarse
	go loadbalancer.VigilarSalud(reg, 5*time.Second)

	ctrl := loadbalancer.NuevoControlador(reg)
	handler := middleware.ConLog("loadbalancer", middleware.ConNodo("loadbalancer", ctrl.Rutas()))

	puerto := config.PuertoHTTP("8081")
	log.Printf("[INFO] Loadbalancer escuchando en %s", puerto)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El loadbalancer se detuvo: %v", err)
	}
}
