package main

import (
	"log"
	"time"

	"logica/internal/basededatos"
	"logica/internal/config"
	"logica/internal/historial"
	"logica/internal/modulos"
	"logica/internal/modulos/tablas"
	"logica/internal/rpc"
	"logica/internal/servidor"
)

// Puerto de la primera instancia si no se define PUERTO_BASE
const puertoPorOmision = 9011

func main() {
	dsn, err := config.DSN()
	if err != nil {
		log.Fatalf("[ERROR] Configuración inválida de la DB %v", err)
	}

	db, err := basededatos.Abrir(dsn, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("[ERROR] No se logró establecer la conexión con POstgres: %v", err)
	}
	defer db.Close()

	log.Printf("[INFO] Módulo %s conectado a la db", tablas.Nombre)

	op := servidor.Opciones{
		Modulo:     tablas.Nombre,
		Instancias: config.NumInstancias(),
		PuertoBase: config.PuertoBase(puertoPorOmision),
		Host:       config.HostAnunciado("localhost"),
		Consul:     config.DireccionConsul("localhost:8500"),
	}

	repo := historial.NuevoRepo(db)

	// Conectar el historial a kafka
	eventos, err := historial.ConectarKafka(config.BrokerKafka("localhost:9092"), tablas.Nombre, repo)
	if err != nil {
		log.Fatalf("[ERROR] No se pudo conectar a Kafka: %v", err)
	}
	defer eventos.Close()

	// Un controlador por instancia, aunque todos comparten el repo y el productor
	nuevo := func(instancia string) rpc.ModuloServer {
		return modulos.NuevoControlador(tablas.Nombre, instancia, repo, eventos, tablas.Resolver)
	}

	log.Printf("[INFO] Módulo %s: %d instancias en el puerto %d", op.Modulo, op.Instancias, op.PuertoBase)

	if err := servidor.Arrancar(op, nuevo); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
