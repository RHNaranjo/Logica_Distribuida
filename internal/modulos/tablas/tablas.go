package tablas

import (
	"encoding/json"
	"fmt"

	"logica/internal/modulos"
)

const Nombre = "tablas"

// Recibir el JSON, después será tablas de Verdad
func Resolver(entrada json.RawMessage) (any, error) {
	return nil, fmt.Errorf("Tablas de verdad...", modulos.ErrNoImplementado)
}
