package formula

import (
	"fmt"
	"slices"
)

// Formula --> nodo
type Formula interface {
	esFormula() // Para que sólo lo que yo quiero pueda ser una fórmula
}

// Estructuras con los tipos de datos que se esperan recibir
type (
	Variable      struct{ Nombre string }
	No            struct{ Modificada Formula }
	Y             struct{ Antecedente, Consequente Formula }
	O             struct{ Antecedente, Consequente Formula }
	Condicional   struct{ Antecedente, Consequente Formula }
	Bicondicional struct{ Antecedente, Consequente Formula }
	Necesario     struct{ Modificada Formula }
	Posible       struct{ Modificada Formula }
)

// Métodos privados para limitar
func (Variable) esFormula()      {}
func (No) esFormula()            {}
func (Y) esFormula()             {}
func (O) esFormula()             {}
func (Condicional) esFormula()   {}
func (Bicondicional) esFormula() {}
func (Necesario) esFormula()     {}
func (Posible) esFormula()       {}

// Devuelve nombres en orden alfabetico
func Variables(f Formula) []string {
	vistas := make(map[string]bool)
	recolectar(f, vistas)

	nombres := make([]string, 0, len(vistas))
	for nombre := range vistas {
		nombres = append(nombres, nombre)
	}

	slices.Sort(nombres)
	return nombres
}

func recolectar(f Formula, vistas map[string]bool) {
	switch n := f.(type) {
	case Variable:
		vistas[n.Nombre] = true
	case No:
		recolectar(n.Modificada, vistas)
	case Necesario:
		recolectar(n.Modificada, vistas)
	case Posible:
		recolectar(n.Modificada, vistas)
	case Y:
		recolectar(n.Antecedente, vistas)
		recolectar(n.Consequente, vistas)
	case O:
		recolectar(n.Antecedente, vistas)
		recolectar(n.Consequente, vistas)
	case Condicional:
		recolectar(n.Antecedente, vistas)
		recolectar(n.Consequente, vistas)
	case Bicondicional:
		recolectar(n.Antecedente, vistas)
		recolectar(n.Consequente, vistas)
	}
}

// Vuelve a texto con los paréntesis para ver qué se interpretó
func Escribir(f Formula) string {
	switch n := f.(type) {
	case Variable:
		return n.Nombre
	case No:
		return "~" + Escribir(n.Modificada)
	case Necesario:
		return "[]" + Escribir(n.Modificada)
	case Posible:
		return "<>" + Escribir(n.Modificada)
	case Y:
		return fmt.Sprintf("(%s & %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case O:
		return fmt.Sprintf("(%s | %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case Condicional:
		return fmt.Sprintf("(%s -> %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	case Bicondicional:
		return fmt.Sprintf("(%s <-> %s)", Escribir(n.Antecedente), Escribir(n.Consequente))
	}

	// Default
	return "?"
}
