package deduccion

import (
	"encoding/json"
	"fmt"

	"logica/internal/modulos"
)

const Nombre = "deduccion"

func Resolver(entrada json.RawMessage) (any, error) {
	return nil, fmt.Errorf("Deducción natural para después", modulos.ErrNoImplementado)
}
