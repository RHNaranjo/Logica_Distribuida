package predicados

import "fmt"

type Formula interface {
	esFormula()
}

type (
	Predicado     struct{ Nombre, Argumento string }
	No            struct{ Modificada Formula }
	Y             struct{ Antecedente, Consequente Formula }
	O             struct{ Antecedente, Consequente Formula }
	Condicional   struct{ Antecedente, Consequente Formula }
	Bicondicional struct{ Antecedente, Consequente Formula }

	// Cuantificadores
	ParaTodo struct {
		Variable string
		Cuerpo   Formula
	}
	Existe struct {
		Variable string
		Cuerpo   Formula
	}
)

func (Predicado) esFormula()     {}
func (No) esFormula()            {}
func (Y) esFormula()             {}
func (O) esFormula()             {}
func (Condicional) esFormula()   {}
func (Bicondicional) esFormula() {}
func (ParaTodo) esFormula()      {}
func (Existe) esFormula()        {}

// Conviertir a texto
func Escribir(f Formula) string {
	switch n := f.(type) {
	case Predicado:
		return n.Nombre + n.Argumento
	case No:
		return "~" + Escribir(n.Modificada)
	case Y:
		return fmt.Sprintf("(%s & %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case O:
		return fmt.Sprintf("(%s | %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case Condicional:
		return fmt.Sprintf("(%s -> %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case Bicondicional:
		return fmt.Sprintf("(%s <-> %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case ParaTodo:
		return fmt.Sprintf("∀%s%s", n.Variable, Escribir(n.Cuerpo))
	case Existe:
		return fmt.Sprintf("∃%s%s", n.Variable, Escribir(n.Cuerpo))
	}

	return "?"
}
