package servidor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"logica/internal/middleware"
)

type Opciones struct {
	Modulo        string // mundos, tablas o deducción natural
	Instancias    int
	PuertoBase    int
	HostAnunciado string
	URLRegistro   string
}

// Anuncio que se envía al LB por instancia
type anuncio struct {
	Modulo    string `json:"modulo"`
	Instancia string `json:"instancia"`
	URL       string `json:"url"`
}

// Levanta N servidores HTTP (según gorutinas). Se bloquea hasta que alguno se caiga
func Arrancar(op Opciones, rutas http.Handler) error {
	errores := make(chan error, op.Instancias)

	for i := 0; i < op.Instancias; i++ {
		puerto := op.PuertoBase + i
		nombre := fmt.Sprintf("%s-%d", op.Modulo, i)
		url := fmt.Sprintf("http://%s:%d", op.HostAnunciado, puerto)

		// ConNodo necesita el nombre de cada instancia particular
		handler := middleware.ConLog(nombre, middleware.ConNodo(nombre, rutas))

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", puerto),
			Handler: handler,

			// Para que las gorutinas no se queden ocupadas infinitamente
			ReadHeaderTimeout: 5 * time.Second,
		}

		go func() {
			log.Printf("[INFO] %s escuchando en %s", nombre, srv.Addr)
			errores <- fmt.Errorf("la instancia %s se detuvo: %w", nombre, srv.ListenAndServe())
		}()

		go registrar(op.URLRegistro, anuncio{
			Modulo:    op.Modulo,
			Instancia: nombre,
			URL:       url,
		})
	}

	// Si algo pasa, se suicida con mucho ruido jajaja
	return <-errores
}

// Avisa al balanceador que la instancia existe
func registrar(urlRegistro string, a anuncio) {
	if urlRegistro == "" {
		log.Printf("[AVISO] %s no se registra: falta REGISTRO_URL", a.Instancia)
		return
	}

	cuerpo, err := json.Marshal(a)
	if err != nil {
		log.Printf("[ERROR] No se pudo serializar el anuncio de %s: %v", a.Instancia, err)
		return
	}

	// Timeout para que no se rompa si el balanceador acepta la conexión y nunca contesta
	cliente := &http.Client{Timeout: 3 * time.Second}

	for intento := 1; intento <= 10; intento++ {
		resp, err := cliente.Post(urlRegistro, "application/json", bytes.NewReader(cuerpo))

		if err == nil {
			codigo := resp.StatusCode
			resp.Body.Close()

			if codigo < 300 {
				log.Printf("[INFO] %s registrada en %s", a.Instancia, urlRegistro)
				return
			}

			err = fmt.Errorf("El balanceador respondió: %d", codigo)
		}

		log.Printf("[AVISO] Intento %d/10 de registrar %s: %v", intento, a.Instancia, err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("[ERROR] %s no pudo registrarse; el balanceador no la verá", a.Instancia)
}
