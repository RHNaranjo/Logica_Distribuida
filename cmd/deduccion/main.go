package main

import (
	"log"

	"logica/internal/config"
	"logica/internal/modulos/deduccion"
	"logica/internal/servidor"
)

// Puerto de la primera instancia si no se define PUERTO_BASE
const puertoPorOmision = 9021

func main() {
	// Creando al servidor
	op := servidor.Opciones{
		Modulo:        "deduccion",
		Instancias:    config.NumInstancias(),
		PuertoBase:    config.PuertoBase(puertoPorOmision),
		HostAnunciado: config.HostAnunciado("localhost"),
		URLRegistro:   config.URLRegistro(),
	}

	ctrl := deduccion.NuevoControlador()

	log.Printf("[INFO] Módulo %s: %d instancias desde el puerto %d", op.Modulo, op.Instancias, op.PuertoBase)

	if err := servidor.Arrancar(op, ctrl.Rutas()); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
