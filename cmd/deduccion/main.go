package main

import (
	"log"
	"time"

	"logica/internal/basededatos"
	"logica/internal/config"
	"logica/internal/historial"
	"logica/internal/modulos"
	"logica/internal/modulos/deduccion"
	"logica/internal/rpc"
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
		Modulo:     deduccion.Nombre,
		Instancias: config.NumInstancias(),
		PuertoBase: config.PuertoBase(puertoPorOmision),
		Host:       config.HostAnunciado("localhost"),
		Consul:     config.DireccionConsul("localhost:8500"),
	}

	repo := historial.NuevoRepo(db)

	// Conectar el historial a kafka
	eventos, err := historial.ConectarKafka(config.BrokerKafka("localhost:9092"), deduccion.Nombre, repo)
	if err != nil {
		log.Fatalf("[ERROR] No se pudo conectar a Kafka: %v", err)
	}
	defer eventos.Close()

	// Un controlador por instancia; todos comparten el repo y el productor
	nuevo := func(instancia string) rpc.ModuloServer {
		return modulos.NuevoControlador(deduccion.Nombre, instancia, repo, eventos, deduccion.Resolver)
	}

	log.Printf("[INFO] Módulo %s: %d instancias en el puerto %d", op.Modulo, op.Instancias, op.PuertoBase)

	if err := servidor.Arrancar(op, nuevo); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
