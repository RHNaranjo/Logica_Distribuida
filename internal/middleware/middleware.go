package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// Agregar el header a cada respuesta para saber quién fue el que respondió
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

		// El id se pone en automático con la función de 'ConPeticion'
		peticion := r.Header.Get("X-Peticion")

		// En caso de que el log no haya llegado por ahí
		if peticion == "" {
			peticion = "-"
		}

		// [tablas-0] p8 GET /salud (1.23ms)
		log.Printf("[%s] %s %s %s (%s)", nodo, peticion, r.Method, r.URL.Path, time.Since(inicio))
	})
}

// Permitir que el front consuma la API
func ConCORS(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Qué páginas pueden leer una respuesta
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Qué métodos de HTTP se pueden usar
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Qué headers se pueden mandar a la petición
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Para que JS no pueda leer los headers aunque lleguen
		w.Header().Set("Access-Control-Expose-Headers", "X-Peticion, X-Instancia, X-Nodo")

		// El navegador pregunta permiso
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		siguiente.ServeHTTP(w, r)
	})
}

// Identificador único a cada petición que entra
func ConPeticion(siguiente http.Handler) http.Handler {
	var contador atomic.Uint64

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("p%d", contador.Add(1))

		// Viaja al LB y de ahí a la instancia
		r.Header.Set("X-Peticion", id)

		// En la respuesta, para buscarlo en los logs
		w.Header().Set("X-Peticion", id)

		siguiente.ServeHTTP(w, r)
	})
}
