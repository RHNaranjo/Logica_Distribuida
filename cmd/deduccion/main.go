package main

import (
	"log"
	"time"

	"logica/internal/basededatos"
	"logica/internal/config"
	"logica/internal/historial"
	"logica/internal/modulos"
	"logica/internal/modulos/deduccion"
	"logica/internal/servidor"
)

// Puerto de la primera instancia si no se define PUERTO_BASE
const puertoPorOmision = 9021

func main() {
	dsn, err := config.DSN()
	if err != nil {
		log.Fatalf("[ERROR] Configuración de base de datos inválida: %v", err)
	}

	// Las tres instancias comparten este pool: *sql.DB es seguro entre goroutines
	db, err := basededatos.Abrir(dsn, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("[ERROR] No se pudo conectar a Postgres: %v", err)
	}
	defer db.Close()

	log.Printf("[INFO] Módulo %s conectado a la base", deduccion.Nombre)

	op := servidor.Opciones{
		Modulo:        deduccion.Nombre,
		Instancias:    config.NumInstancias(),
		PuertoBase:    config.PuertoBase(puertoPorOmision),
		HostAnunciado: config.HostAnunciado("localhost"),
		URLRegistro:   config.URLRegistro(),
	}

	repo := historial.NuevoRepo(db)
	ctrl := modulos.NuevoControlador(deduccion.Nombre, repo, deduccion.Resolver)

	log.Printf("[INFO] Módulo %s: %d instancias desde el puerto %d", op.Modulo, op.Instancias, op.PuertoBase)

	if err := servidor.Arrancar(op, ctrl.Rutas()); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
