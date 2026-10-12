package main

import (
	"log"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"logica/internal/config"
	"logica/internal/middleware"
	"logica/internal/puerta"
)

func main() {
	destino := config.DestinoGRPC("localhost:9000")

	// Preparar la conexión gRPC con nginx
	conexion, err := grpc.NewClient(destino, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[ERROR] GRPC_DESTINO inválido %q: %v", destino, err)
	}
	defer conexion.Close()

	ctrl := puerta.NuevoControlador(conexion)

	// Peticion(Log(Nodo(rutas)))
	var handler http.Handler = ctrl.Rutas()
	handler = middleware.ConNodo("middleware", handler) // Sólo marcar respuesta
	handler = middleware.ConLog("middleware", handler)
	handler = middleware.ConPeticion(handler) // Después de ConLog para que el ID exista

	puerto := config.PuertoHTTP("8080")
	log.Printf("[INFO] middleware escuchando en %s. Llama por gRPC a %s", puerto, destino)

	if err := http.ListenAndServe(puerto, handler); err != nil {
		log.Fatalf("[ERROR] El middleware se detuvo: %v", err)
	}
}
