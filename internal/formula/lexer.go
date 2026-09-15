package formula

import (
	"fmt"
	"unicode"
)

type tipoToken int

// lista de tokens
const (
	tokenVariable tipoToken = iota
	tokenNo
	tokenY
	tokenO
	tokenCondicional
	tokenBicondicional
	tokenNecesario
	tokenPosible
	tokenParentIzq
	tokenParentDer
	tokenFin
)

type token struct {
	tipo  tipoToken
	texto string
}

// Solo ASCII
func esLetra(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func esDigito(r rune) bool {
	return r >= '0' && r <= '9'
}

// Convertir el texto en una lista
func escanear(entrada string) ([]token, error) {
	// Convertir a caracteres únicos
	runas := []rune(entrada)

	// Espacio para los tokens
	tokens := make([]token, 0, len(runas))
	i := 0

	// Recorrer cada caracter
	for i < len(runas) {
		r := runas[i]

		// Analizar cada caso
		switch {
		// Si es espacio, se lo brinca
		case unicode.IsSpace(r):
			i++

		// Si es letra...
		case esLetra(r):
			// Obtiene el inicio actual y avanza i
			inicio := i
			i++

			// Si después viene otro dígito, se lo brinca
			for i < len(runas) && (esLetra(runas[i]) || esDigito(runas[i])) {
				i++
			}

			// Guardar en el token todo lo que venga ahí ('p', 'q', 'Miami')
			tokens = append(tokens, token{tokenVariable, string(runas[inicio:i])})

		case r == '~':
			tokens = append(tokens, token{tokenNo, "~"})
			i++

		case r == '&':
			tokens = append(tokens, token{tokenY, "&"})
			i++

		case r == '|':
			tokens = append(tokens, token{tokenO, "|"})
			i++

		case r == '(':
			tokens = append(tokens, token{tokenParentIzq, "("})
			i++

		case r == ')':
			tokens = append(tokens, token{tokenParentDer, ")"})
			i++

		// Necesario []
		case r == '[':
			if i+1 >= len(runas) || runas[i+1] != ']' {
				return nil, fmt.Errorf("[ERROR] Se esperaba un ']' después de '['")
			}
			tokens = append(tokens, token{tokenNecesario, "[]"})
			i += 2

		// Posible <> o bicondicional <->
		case r == '<':
			switch {
			case i+1 < len(runas) && runas[i+1] == '>':
				tokens = append(tokens, token{tokenPosible, "<>"})
				i += 2

			case i+2 < len(runas) && runas[i+1] == '-' && runas[i+2] == '>':
				tokens = append(tokens, token{tokenBicondicional, "<->"})
				i += 3

			default:
				return nil, fmt.Errorf("[ERROR] Se esperaba '<>' o '<->'")
			}

		// Condicional ->
		case r == '-':
			if i+1 >= len(runas) || runas[i+1] != '>' {
				return nil, fmt.Errorf("[ERROR] Se esperaba ->")
			}
			tokens = append(tokens, token{tokenCondicional, "->"})
			i += 2

		default:
			return nil, fmt.Errorf("No se pudo leer el token recibido: %q", string(r))
		}
	}

	tokens = append(tokens, token{tokenFin, ""})
	return tokens, nil
}
