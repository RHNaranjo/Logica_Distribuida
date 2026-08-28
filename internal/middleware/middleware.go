package middleware

import (
	"log"
	"net/http"
	"time"
)

// Agregar el header a cada respuesta para saber quién fue el que respondió (demostrar cómputo distribuido)
func ConNodo(nodo string, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// El header se escribe antes de delegar
		// Todo lo que viene después se ignora
		w.Header().Set("X-Nodo", nodo)
		siguiente.ServeHTTP(w, r)
	})
}

// Registrar peticiones y su tardanza
func ConLog(nodo string, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		// Medir cuando el handler ya terminó
		siguiente.ServeHTTP(w, r)

		// Odio escribir estos números así AJAJA
		log.Printf("[%s] %s %s (%S)", nodo, r.Method, r.URL.Path, time.Since(inicio))
	})
}
