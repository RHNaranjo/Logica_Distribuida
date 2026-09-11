package web

import (
	"encoding/json"
	"log"
	"net/http"
)

type Error struct {
	Mensaje string `json:"error"`
}

// Escribe cualquier valor como respuesta JSON
func JSON(w http.ResponseWriter, codigo int, datos any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)

	if err := json.NewEncoder(w).Encode(datos); err != nil {
		log.Printf("[ERROR] No se pudo escribir la respuesta: %v", err)
	}
}

func Fallo(w http.ResponseWriter, codigo int, mensaje string) {
	JSON(w, codigo, Error{Mensaje: mensaje})
}
