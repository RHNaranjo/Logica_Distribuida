package kripke

import (
	"fmt"
	"sort"
)

// Relaciones entre mundos posibles
type Relacion struct {
	Origen  int
	Destino int
}

// Variables verdaderas y falsas en un mundo posible
type Valuacion struct {
	Verdaderas []string
	Falsas     []string
}

// Modelo de Kropke = <Mundos, Relaciones, Variables>
type Modelo struct {
	NumMundos   int
	Relaciones  []Relacion
	Valuaciones map[int]Valuacion
}

// Mundos accesibles desde el mundo dado
func (m Modelo) Sucesores(mundo int) []int {
	vistos := make(map[int]bool)

	var sucesores []int

	// Recorre cada relación que tiene este mundo para crear una lista
	for _, r := range m.Relaciones {
		if r.Origen == mundo && !vistos[r.Destino] {
			vistos[r.Destino] = true
			sucesores = append(sucesores, r.Destino)
		}
	}

	sort.Ints(sucesores)

	return sucesores
}

// Devuelve valuaciones del mundo
func (m Modelo) ValuacionDe(mundo int) Valuacion {
	return m.Valuaciones[mundo]
}

// Revisa si una variable es verdadera en un mundo
func (v Valuacion) EsVerdadera(variable string) bool {
	for _, nombre := range v.Verdaderas {
		if nombre == variable {
			return true
		}
	}

	return false
}

// Validar que el modelo tenga sentido
func (m Modelo) Validar() error {
	if m.NumMundos <= 0 {
		return fmt.Errorf("El número de mundos debe ser mayor a 0, pero se recibió %d", m.NumMundos)
	}

	ultimo := m.NumMundos - 1 // Se mpieza a contar en 0

	// Recorrer cada relacion y ver que incluya mundos reales
	for _, r := range m.Relaciones {
		if r.Origen < 0 || r.Origen > ultimo {
			return fmt.Errorf("La relación <%d,%d> tiene un origen fuera del rango de mundos válidos (0...%d)", r.Origen, r.Destino, ultimo)
		}
		if r.Destino < 0 || r.Destino > ultimo {
			return fmt.Errorf("La relacion <%d,%d> tiene un destino fuera del rango de mundos válidos (0...%d)", r.Origen, r.Destino, ultimo)
		}
	}

	// Revisar que todas las valuaciones sean para mundos reales
	for mundo, v := range m.Valuaciones {
		if mundo < 0 || mundo > ultimo {
			return fmt.Errorf("Hay una valuación para el mundo %d, que no existe. Los mundos válidos son: 0...%d", mundo, ultimo)
		}

		if err := v.validarCoherencia(mundo); err != nil {
			return err
		}
	}

	// Todo correcto
	return nil
}

// Revisar que ninguna variable aparezca en verdaderas y en falsas
func (v Valuacion) validarCoherencia(mundo int) error {
	veraderas := make(map[string]bool, len(v.Verdaderas))

	for _, nombre := range v.Verdaderas {
		veraderas[nombre] = true
	}

	for _, nombre := range v.Falsas {
		if veraderas[nombre] {
			return fmt.Errorf("La variable %q está declarada verdadera y falsa a la vez en un mismo mundo: %d.", nombre, mundo)
		}
	}

	return nil
}
