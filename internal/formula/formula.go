package formula

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
