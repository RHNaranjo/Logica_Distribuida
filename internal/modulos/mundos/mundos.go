package mundos

import (
	"encoding/json"
	"fmt"

	"logica/internal/modulos"
)

const Nombre = "mundos"

func Resolver(entrada json.RawMessage) (any, error) {
	return nil, fmt.Errorf("Después: mundos posibles", modulos.ErrNoImplementado)
}
