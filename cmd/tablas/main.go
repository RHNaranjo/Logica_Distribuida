package main

import (
	"log"
	"time"

	"logica/internal/basededatos"
	"logica/internal/config"
	"logica/internal/historial"
	"logica/internal/modulos"
	"logica/internal/modulos/tablas"
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
		Modulo:        tablas.Nombre,
		Instancias:    config.NumInstancias(),
		PuertoBase:    config.PuertoBase(puertoPorOmision),
		HostAnunciado: config.HostAnunciado("localhost"),
		URLRegistro:   config.URLRegistro(),
	}

	repo := historial.NuevoRepo(db)
	ctrl := modulos.NuevoControlador(tablas.Nombre, repo, tablas.Resolver)

	log.Printf("[INFO] Módulo %s: %d instancias en el puerto %d", op.Modulo, op.Instancias, op.PuertoBase)

	if err := servidor.Arrancar(op, ctrl.Rutas()); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
