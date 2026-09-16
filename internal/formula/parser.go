package formula

import "fmt"

type parser struct {
	tokens []token 
	pos int 
}

func (p *parser) tokenActual() token {
	return p.tokens[p.pos]
}

func (p *parser) avanzarToken() token {
	t := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return t 
}

// Por aquí entra el paquete 
func Parsear(entrada string) (Formula, error) {
	tokens, err := escanear(entrada)

	if err != nil {
		return nil, err 
	}

	// Si no hay nada más que leer +
	if len(tokens) == 1 {
		return nil, fmt.Errorf("[ERROR] La fórmula está vacía")
	}

	p := &parser{tokens}

	f, err := p.bicondicional() 
	if err != nil {
		return nil, err 
	}

	// Si algo sobró, la fórmula estaba mal 
	if p.tokenActual().tipo != tokenFin {
		return nil, fmt.Errorf("[ERROR] Sobra %q después de la fórmula", p.actual().texto)
	}

	return f, nil 
}

// <->
func (p *parser) bicondicional() (Formula, error) {
	ant, err := p.condicional()
	if err != nil {
		return nil, err 
	}

	for p.tokenActual().tipo == tokenBicondicional {
		p.avanzarToken()

		cons, err := p.condicional()
		if err != nil {
			return nil, err 
		}

		cons = Bicondicional{ant, const}
	}

	return ant, nil
}

// El condicional se asocia a la derecha: p -> q -> r ==> p -> (q -> r)
func (p *parser) condicional() (Formula, error) {
	ant, err := p.disyuncion()
	if err != nil {
		return nil, err 
	}

	if p.tokenActual().tipo == tokenCondicional {
		p.avanzarToken()

		cons, err := p.condicional()
		if err != nil {
			return nil, err 
		}

		return Condicional{ant, cons}
	}

	return ant, nil 
}

// Disyuncion, misma recursividad de la conjunción 
func (p *parser) disyuncion() (Formula, error) {
	ant, err := p.conjuncion()

	if err != nil {
		return nil, err 
	}

	for p.tokenActual().tipo == tokenO {
		p.avanzarToken()

		cons, err := p.conjuncion()
		if err != nil {
			return nil, err 
		}

		ant = O{ant, cons}
	}

	return ant, nil 
}

// Conjunción 
func (p *parser) conjuncion() (Formula, error) {
	ant, err := p.modBasicos()

	if err != nil {
		return nil, err 
	}

	// Recursividad p & q & r -> p & (q & r)
	for p.tokenActual().tipo == tokenY {
		p.avanzarToken()

		cons, err := p.modBasicos()
		if err != nil {
			return nil, err 
		}

		ant = Y{ant, cons}
	}

	return ant, nil
}

// ~, [], <>, que sólo modifican la variable a su derecha 
// Se llaman a sí mismos para formar ~~p, ~[]p, []<>~p
func (p *parser) modBasicos() (Formula, error) {
	switch p.tokenActual().tipo {
	case tokenNo:
		p.avanzarToken()
		modificada, err := p.modBasicos()
		
		if err != nil {
			return nil, err 
		}

		return No{modificada}, nil 

	case tokenNecesario:
		p.avanzarToken()
		modificada, err := p.modBasicos()

		if err != nil {
			return nil, err 
		}

		return Necesario{modificada}, nil 

	case tokenPosible:
		p.avanzarToken()
		modificada, err := p.modBasicos()

		if err != nil {
			return nil, err 
		}

		return Posible{modificada}, nil 
	}

	return p.atomo()
}

// Variable o fórmula entre paréntesis 
func (p *parser) atomo() (Formula, error) {
	t := p.tokenActual()

	switch t.tipo {
	case tokenVariable:
		p.avanzarToken()
		return Variable{Nombre: t.texto}, nil 

	case tokenParentIzq:
		p.avanzarToken()
		
		dentro, err := p.bicondicional()
		if err != nil {
			return nil, err 
		}

		if p.tokenActual().tipo != tokenParentDer {
			return nil, fmt.Errorf("[ERROR] Falta cerrar un paréntesis")
		}
		p.avanzarToken()

		return dentro, nil 

	case tokenFin:
		return nil, fmt.Errorf("[ERROR] La fórmula terminó antes de tiempo")
	}

	return nil, fmt.Errorf("[ERROR] No se esperaba un %q aquí", t.texto)
}

