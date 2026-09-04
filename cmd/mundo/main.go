package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"mundos_posibles/internal/basededatos"
	"mundos_posibles/internal/config"
	"mundos_posibles/internal/middleware"
	"mundos_posibles/internal/mundo"
)

func main() {
	// Crear el mundo posible
	id, err := config.MundoID()
	if err != nil {
		log.Fatalf("[ERROR] Configuración inválida: %v", err)
	}

	dsn, err := config.DSN()
	if err != nil {
		log.Fatalf("[ERROR] Configuración de base de datos inválida: %v", err)
	}

	db, err := basededatos.Abrir(dsn, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("[ERROR] No se pudo conectar a Postgres: %v", err)
	}
	defer db.Close()

	nombre := fmt.Sprintf("mundo%d", id)
	log.Printf("[INFO] %s conectado a su base de datos", nombre)

	repo := mundo.NuevoRepo(db)
	ctrl := mundo.NuevoControlador(repo, id)

	handler := middleware.ConLog(nombre, middleware.ConNodo(nombre, ctrl.Rutas()))

	puerto := config.PuertoHTTP()
	log.Printf("[INFO] %s escuchando en %s", nombre, puerto)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El nodo %s se detuvo: %v", nombre, err)
	}
}
