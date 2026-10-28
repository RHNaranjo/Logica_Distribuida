package tablas

import (
	"encoding/json"
	"fmt"

	"logica/internal/formula"
)

const Nombre = "tablas"

// Lo que manda el frontent
type Entrada struct {
	Formula string `json:"formula"`
}

// Leer la fórmula para construir y clasificar la TV
func Resolver(entrada json.RawMessage) (any, error) {
	var e Entrada

	// Leer lo que se mandó
	if err := json.Unmarshal(entrada, &e); err != nil {
		return nil, fmt.Errorf("la entrada debe ser {\"formula\": \"...\": %w}", err)
	}

	// Convertir el texto en un árbol
	f, err := formula.Parsear(e.Formula)
	if err != nil {
		return nil, err
	}

	// Generarl los renglones
	return formula.ConstruirTabla(f)
}
