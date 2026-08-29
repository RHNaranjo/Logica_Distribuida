package kripke

import (
	"fmt"
	"strconv"
	"strings"
)

// Construir y validar el modelo a partir de mundos, relaciones y valuaciones
func ConstruirModelo(numMundos, relaciones, valuaciones string) (Modelo, error) {
	n, err := ParsearNumMundos(numMundos)
	if err != nil {
		return Modelo{}, err
	}

	r, err := ParsearRelaciones(relaciones)
	if err != nil {
		return Modelo{}, err
	}

	v, err := ParsearValuaciones(valuaciones)
	if err != nil {
		return Modelo{}, err
	}

	m := Modelo{
		NumMundos:   n,
		Relaciones:  r,
		Valuaciones: v,
	}

	if err := m.Validar(); err != nil {
		return Modelo{}, err
	}

	return m, nil
}

// Convertir string en int
func ParsearNumMundos(mundos string) (int, error) {
	// Eliminar espacios en torno a la palabra
	mundos = strings.TrimSpace(mundos)

	if mundos == "" {
		return 0, fmt.Errorf("el número de mundos no puede estar vacío")
	}

	n, err := strconv.Atoi(mundos)

	if err != nil {
		return 0, fmt.Errorf("el número de mundos debe ser un entero, pero se recibió %q: %w", mundos, err)
	}

	if n <= 0 {
		return 0, fmt.Errorf("el número de mundos debe ser positivo, se recibió %d", n)
	}

	return n, nil
}

// Convertir el <0,0> en una []Relacion. Mucho de lo que viene aquí no será necesario porque el frontend enviará todo en el formato correcto al middleware
func ParsearRelaciones(rel string) ([]Relacion, error) {
	rel = strings.TrimSpace(rel)

	// No hay problema si
	if rel == "" {
		return nil, nil
	}

	var relaciones []Relacion

	partes := strings.Split(rel, ">") // Sólo me quedo con '<0,1'

	for i, parte := range partes {
		// Eliminar la última parte (espacios vacíos)
		if i == len(partes)-1 {
			if resto := strings.Trim(parte, " \t,"); resto != "" {
				return nil, fmt.Errorf("texto no reconocido: %q, revisar el formato", resto)
			}

			break
		}

		// Quitar separadores de la iteración anterior
		parte = strings.Trim(parte, " \t,")

		if !strings.HasPrefix(parte, "<") {
			return nil, fmt.Errorf("se esperaba '<' al inicio del par, pero se recibió: %q", parte)
		}

		numeros := strings.Split(strings.TrimPrefix(parte, "<"), ",")

		if len(numeros) != 2 {
			return nil, fmt.Errorf("un par debe tener exactamente dos números separados por coma, se recibió %q>", parte)
		}

		// Obtener el nodo de origen de esta iteración
		origen, err := strconv.Atoi(strings.TrimSpace(numeros[0]))

		if err != nil {
			return nil, fmt.Errorf("origen inválido en %q: %w", parte, err)
		}

		destino, err := strconv.Atoi(strings.TrimSpace(numeros[1]))

		if err != nil {
			return nil, fmt.Errorf("destino inválido en %q: %w", parte, err)
		}

		// Añadir la relación actual a la lista
		relaciones = append(relaciones, Relacion{Origen: origen, Destino: destino})
	}

	return relaciones, nil
}

// Parsear los valores de cada relación. Recibe {0 | p, q}, {1 | ~p, q}, etc...
func ParsearValuaciones(val string) (map[int]Valuacion, error) {
	valuaciones := make(map[int]Valuacion)

	val = strings.TrimSpace(val)
	if val == "" {
		return valuaciones, nil
	}

	partes := strings.Split(val, "}")

	for i, parte := range partes {
		if i == len(partes)-1 {
			if resto := strings.Trim(parte, " \t,"); resto != "" {
				return nil, fmt.Errorf("texto no reconocido: %q, revisar el formato", resto)
			}

			break
		}

		parte = strings.Trim(parte, " \t,")

		if !strings.HasPrefix(parte, "{") {
			return nil, fmt.Errorf("se esperaba '{' al inicio del texto, se recibió: %q", parte)
		}

		// Obtener el cuerpo de los valores del mundo y el mundo al que pertenecen
		cuerpo := strings.TrimPrefix(parte, "{")
		mitades := strings.Split(cuerpo, "|")

		if len(mitades) != 2 {
			return nil, fmt.Errorf("se esperaba 'mundo | variables' dentro del cuerpo, pero se recibió: %q", cuerpo)
		}

		// Validar que existe el mundo
		mundo, err := strconv.Atoi(strings.TrimSpace(mitades[0]))
		if err != nil {
			return nil, fmt.Errorf("número de mundo inválido en %q: %w", cuerpo, err)
		}

		// Ver si el mundo se repite
		if _, repetido := valuaciones[mundo]; repetido {
			return nil, fmt.Errorf("el mundo %d tiene más de una valuación declarada", mundo)
		}

		valuacion, err := parsearVariables(mitades[1], mundo)
		if err != nil {
			return nil, err
		}

		valuaciones[mundo] = valuacion
	}

	return valuaciones, nil
}

// Convertir "p, ~q" en una valuacion
func parsearVariables(entrada string, mundo int) (Valuacion, error) {
	var v Valuacion

	entrada = strings.TrimSpace(entrada)
	if entrada == "" {
		return v, nil
	}

	for _, literal := range strings.Split(entrada, ",") {
		literal = strings.TrimSpace(literal)

		if literal == "" {
			return Valuacion{}, fmt.Errorf("hay una coma de más en la valuación del mundo %d", mundo)
		}

		negada := strings.HasPrefix(literal, "~")
		nombre := strings.TrimSpace(strings.TrimPrefix(literal, "~"))

		if err := validarNombreVariable(nombre, mundo); err != nil {
			return Valuacion{}, err
		}

		if negada {
			v.Falsas = append(v.Falsas, nombre)
		} else {
			v.Verdaderas = append(v.Verdaderas, nombre)
		}
	}

	return v, nil
}

// Aceptar nombres de las variables
func validarNombreVariable(nombre string, mundo int) error {
	if nombre == "" {
		return fmt.Errorf("hay un '~' sin variable en la valuación del mundo %d", mundo)
	}

	for i, r := range nombre {
		esLetra := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		esDigito := r >= '0' && r <= '9'

		if i == 0 && !esLetra {
			return fmt.Errorf("la variable %q del mundo %d debe empezar con una letra", nombre, mundo)
		}

		if !esLetra && !esDigito {
			return fmt.Errorf("la variable %q del mundo %d solo puede contener letras y dígitos", nombre, mundo)
		}
	}

	return nil
}
