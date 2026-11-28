package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"logica/internal/config"
	"logica/internal/middleware"
	"logica/internal/observabilidad"
	"logica/internal/puerta"
)

func main() {
	// Arrancar las trazas antes que todo lo demás
	apagar, err := observabilidad.Iniciar(context.Background(), "middleware", config.NombreInstancia("middleware"))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	defer apagar(context.Background())

	// El DNs de Kube dice lo que hay en cada módulo
	destinos := map[string]string{
		"mundos":    "mundos:9001",
		"tablas":    "tablas:9011",
		"deduccion": "deduccion:9021",
	}

	ctrl, err := puerta.NuevoControlador(destinos)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	// Traza(Peticion(Log(Nodo(rutas))))
	var handler http.Handler = ctrl.Rutas()
	handler = middleware.ConNodo("middleware", handler)
	handler = middleware.ConLog("middleware", handler)
	handler = middleware.ConPeticion(handler) // Después de ConLog para que el ID exista
	handler = otelhttp.NewHandler(handler, "middleware", otelhttp.WithFilter(conTraza))

	puerto := config.PuertoHTTP("8080")
	log.Printf("[INFO] middleware escuchando en %s. Módulos: %v", puerto, destinos)

	// Para que main pueda esperar la señal
	go func() {
		log.Fatalf("[ERROR] El middleware se detuvo: %v", http.ListenAndServe(puerto, handler))
	}()

	// Esperar a que Kubernetes pida apagar, para que el defer mande los últimos spans
	apagado := make(chan os.Signal, 1)
	signal.Notify(apagado, syscall.SIGINT, syscall.SIGTERM)
	<-apagado

	log.Printf("[INFO] Apagando middleware")
}

// Recibe una petición y devuelve true si se traza
func conTraza(r *http.Request) bool {
	return r.URL.Path != "/salud" && r.URL.Path != "/api/estado"
}
