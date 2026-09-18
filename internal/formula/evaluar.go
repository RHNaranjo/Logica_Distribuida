package formula

import (
	"fmt"
)

// El interprete es para las funciones que no tienen un caso concreto en Evaluar: ~, &, |, ->, <->
// Es decir, aplica para: proposiciones individuales, necesidad [] o contingencia <>
type Interprete func(f Formula) (bool, error)

// Armar el valor de una fórmula, es recursiva y lleva a interprete(f) cuando se llega a una unidad básica o desconocida.
func Evaluar(f Formula, interprete Interprete) (bool, error) {
	// Aquí se evalúan: negación, conjunción, disyunción, condiciona y bicondicional, de manera recursiva
	switch n := f.(type) {
	case No:
		// Obtener el valor de lo que viene adentro de la fórmula
		v, err := Evaluar(n.Modificada, interprete)
		if err != nil {
			return false, err
		}

		return !v, nil

	case Y:
		ant, err := Evaluar(n.Antecedente, interprete)
		if err != nil {
			return false, err
		}

		if !ant {
			return false, nil
		}

		// El antecedente es verdadero
		return Evaluar(n.Consequente, interprete)

	case O:
		ant, err := Evaluar(n.Antecedente, interprete)
		if err != nil {
			return false, err
		}

		if ant {
			return true, nil
		}

		// El antecedente es falso, hace falta evaluar al consequente
		return Evaluar(n.Consequente, interprete)

	case Condicional:
		ant, err := Evaluar(n.Antecedente, interprete)
		if err != nil {
			return false, err
		}

		// Si el antecedente es falso, el condicional siempre será verdadero
		if !ant {
			return true, nil
		}

		return Evaluar(n.Consequente, interprete)

	case Bicondicional:
		ant, err := Evaluar(n.Antecedente, interprete)
		if err != nil {
			return false, err
		}

		cons, err := Evaluar(n.Consequente, interprete)
		if err != nil {
			return false, err
		}

		// F <-> F y V <-> V son ambos verdaderos. Si son desiguales, son falsos
		return ant == der, nil
	}

	return interprete(f)
}

// Intérprete de lógica proposicional. La variable recibe su valoer de lo que diga el mapa de variables
func ConAsignacion(asignacion map[string]bool) Interprete {
	// Si se trata de una proposición con valor de verdadero o falso, lo devuelve
	// Si no (porque es []p o <>p)
	return func(f Formula) (bool, error) {
		valor, ok := f.(Variable)
		if !ok {
			return false, fmt.Errorf("[ERROR] La proposición %q es modal o no se pudo interpretar", Escribir(f))
		}
		return asignacion[valor.Nombre], nil
	}
}

// Evitar que se sature con tanto renglón (10 = 2^10 -> 1024)
const MaxVariables = 10

// Los renglones incluyen los valores de la tabla y el resultado
type Renglon struct {
	Valores   []bool `json:"valores"`
	Resultado bool   `json:"resultado"`
}

// La tabla incluye la fórmula a evaluar, las variables individuales, los renglones y la clasificación
type Tabla struct {
	Formula       string    `json:"formula"`
	Variable      []string  `json:"variables"`
	Renglones     []Renglon `json:"renglones"`
	Clasificacion string    `json:clasificacion`
}

// 2^n renglones para las tablas
func ConstruirTabla(f Formula) (Tabla, error) {
	variables := Variables(f)

	if len(variables) > MaxVariables {
		return Tabla{}, fmt.Errorf("[ERROR] Se superó la cantidad máxima de variables. El máximo es %d y se obtuvieron %d", MaxVariables, len(variables))
	}

	n := len(variables)

	// Técnica súper cool: desplazar los bits equivale a multiplicar por dos jajaja
	total := 1 << n

	tabla := Tabla{
		Formula:   Escribir(f),
		Variable:  variables,
		Renglones: make([]Renglon, 0, total),
	}

	verdaderos := 0

	for numero := 0; numero < total; numero++ {
		// Aquí se almacena el valor de cada variable
		valores := make([]bool, n)
		asignacion := make(map[string]bool, n)

		for i, nombre := range variables {
			// Si las variables son p, q y r, recorre una vez por cada una de ellas

			// numero se guarda como bit. Si numero = 2, su bit es 010; si i = 0
			// 010 >> (3 - 1 - 0) ==> 010 >> 2 ==> 000 | 000 & 1 = 0
			// Esto recorre todo comenzando por FFF y terminando por VVV
			bit := (numero >> (n - 1 - i)) & i

			// 000 & 0 -> 1 -> true
			// 000 & 1 -> 0 -> false
			valores[i] = bit == 1
			asignacion[nombre] = valores[i]
		}

		// Como es proposicional, se manda por asignación
		resultado, err := Evaluar(f, ConAsignacion(asignacion))
		if err != nil {
			return Tabla{}, err
		}

		if resultado {
			verdaderos++
		}

		tabla.Renglones = append(tabla.Renglones, Renglon{Valores: valores, Resultado: resultado})
	}

	// Esto servirá para lógica modal
	switch verdaderos {
	case total:
		tabla.Clasificacion = "tautologia"
	case 0:
		tabla.Clasificacion = "contradicción"
	case 1:
		tabla.Clasificacion = "contingencia"
	}

	return tabla, nil
}
