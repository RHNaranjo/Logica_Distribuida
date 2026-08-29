package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"mundos_posibles/internal/config"
	"mundos_posibles/internal/middleware"
)

type respuestaSalud struct {
	MundoID     int  `json:"mundo_id"` // Se llama mundo_id el campo (para que no sea 0/1)
	Configurado bool `json:"configurado"`
}

func main() {
	// Crear el mundo posible
	id, err := config.MundoID()
	if err != nil {
		log.Fatalf("[ERROR] Configuración inválida: %v", err)
	}

	nombre := fmt.Sprintf("mundo%d", id)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		respuesta := respuestaSalud{
			MundoID:     id,
			Configurado: false,
		}

		if err := json.NewEncoder(w).Encode(respuesta); err != nil {
			log.Printf("[ERROR] No se pudo escribir la respuesta de salud: %v", err)
		}
	})

	// Obtener el Handler
	handler := middleware.ConLog(nombre, middleware.ConNodo(nombre, mux))

	// Mostrar el nombre del backend y el puerto
	puerto := config.PuertoHTTP()
	log.Printf("[INFO] %s escuchando en %s", nombre, puerto)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El nodo %s se detuvo: %v", nombre, err)
	}
}
