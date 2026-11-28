package main

import (
	"context"
	"log"
	"time"

	"logica/internal/basededatos"
	"logica/internal/config"
	"logica/internal/historial"
	"logica/internal/modulos"
	"logica/internal/modulos/mundos"
	"logica/internal/observabilidad"
	"logica/internal/rpc"
	"logica/internal/servidor"
)

// Puerto de la primera instancia si no se define PUERTO_BASE
const puertoPorOmision = 9001

func main() {
	// Arrancar las trazas antes que todo lo demás
	apagar, err := observabilidad.Iniciar(context.Background(), mundos.Nombre, config.NombreInstancia(mundos.Nombre))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	defer apagar(context.Background())

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

	log.Printf("[INFO] Módulo %s conectado a la base", mundos.Nombre)

	op := servidor.Opciones{
		Modulo:     mundos.Nombre,
		Nombre:     config.NombreInstancia(mundos.Nombre),
		Instancias: config.NumInstancias(),
		PuertoBase: config.PuertoBase(puertoPorOmision),
	}

	repo := historial.NuevoRepo(db)

	// Conectar el historial a kafka
	eventos, err := historial.ConectarKafka(config.BrokerKafka("localhost:9092"), mundos.Nombre, repo)
	if err != nil {
		log.Fatalf("[ERROR] No se pudo conectar a Kafka: %v", err)
	}
	defer eventos.Close()

	// Un controlador por instancia
	nuevo := func(instancia string) rpc.ModuloServer {
		return modulos.NuevoControlador(mundos.Nombre, instancia, repo, eventos, mundos.Resolver)
	}

	log.Printf("[INFO] Módulo %s: %d instancias en el puerto %d", op.Modulo, op.Instancias, op.PuertoBase)

	if err := servidor.Arrancar(op, nuevo); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
