package formula

import "testing"

func TestEscanear(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado []tipoToken
	}{
		{"p", []tipoToken{tokenVariable, tokenFin}},
		{"p & q", []tipoToken{tokenVariable, tokenY, tokenVariable, tokenFin}},
		{"p->q", []tipoToken{tokenVariable, tokenCondicional, tokenVariable, tokenFin}},
		// El caso delicado: <-> y <> comparten el primer carácter
		{"p<->q", []tipoToken{tokenVariable, tokenBicondicional, tokenVariable, tokenFin}},
		{"<>p", []tipoToken{tokenPosible, tokenVariable, tokenFin}},
		{"[]p", []tipoToken{tokenNecesario, tokenVariable, tokenFin}},
		{"<><->p", []tipoToken{tokenPosible, tokenBicondicional, tokenVariable, tokenFin}},
		{"P1", []tipoToken{tokenVariable, tokenFin}},
	}

	for _, caso := range casos {
		tokens, err := escanear(caso.entrada)
		if err != nil {
			t.Errorf("[ERROR] Escanear (%q) obtuvo el error: %v", caso.entrada, err)
			continue
		}

		if len(tokens) != len(caso.esperado) {
			t.Errorf("[ERROR] Escanear (%q) dio %d tokens, pero se esperaban %d", caso.entrada, len(tokens), len(caso.esperado))
			continue
		}

		for i, tipo := range caso.esperado {
			if tokens[i].tipo != tipo {
				t.Errorf("escanear (%q) token %d: se obtuvo %v, pero se esperaba obtener %v", caso.entrada, i, tokens[i].tipo, tipo)
			}
		}
	}
}

func TestEscanearErrores(t *testing.T) {
	malas := []string{"p $ q", "p - q", "[p]", "p < q", "p ∀ q"}

	for _, entrada := range malas {
		if _, err := escanear(entrada); err == nil {
			t.Errorf("[ERROR] Escanear(%q) tenía que fallar, pero no falló", entrada)
		}
	}
}

func TestParser(t *testing.T) {
	casos := []struct{ entrada, esperado string }{
		// No debería quedar ~(p & q)
		{"~p & q", "(~p & q)"},

		// La conjunción tiene preferencia sobre la disyunción
		{"p | q & r", "(p | (q & r))"},
		{"p | q -> r", "((p | q) -> r)"},

		// Toma prioridad el condicional, que asocia hacia la derecha
		{"p -> q <-> r", "((p -> q) <-> r)"},
		{"p -> q -> r", "(p -> (q -> r))"},

		// Recursividad en la conjunción y en la disyunción
		{"p & q & r", "((p & q) & r)"},
		{"p | q | r", "((p | q) | r)"},

		// El bicondicional da igual
		{"p <-> q <-> r", "((p <-> q) <-> r)"},

		// Victoria de los paréntesis
		{"(p | q) & r", "((p | q) & r)"},

		// Modificadores modales encadenados
		{"[]<>p", "[]<>p"},
		{"~[]~p", "~[]~p"},
		{"[](p -> q)", "[](p -> q)"},

		// Probar con proposiciones compuestas
		{"P1 & Q2", "(P1 & Q2)"},
		{"A <-> Z", "(A <-> Z)"},
	}

	for _, caso := range casos {
		f, err := Parsear(caso.entrada)
		if err != nil {
			t.Errorf("[ERROR] Parsear(%q) dio error: %v", caso.entrada, err)
			continue
		}

		if obtenido := Escribir(f); obtenido != caso.esperado {
			t.Errorf("[ERROR] Parsear(%q) = %q, pero se esperaba %q", caso.entrada, obtenido, caso.esperado)
		}
	}
}

func TestParsearErrores(t *testing.T) {
	malas := []string{
		"",       // vacía
		"p &",    // falta el operando derecho
		"& p",    // falta el izquierdo
		"(p & q", // paréntesis sin cerrar
		"p & q)", // paréntesis de más
		"p q",    // dos variables juntas
		"~",      // negación sin nada
		"p <-> ", // bicondicional sin operando
	}

	for _, entrada := range malas {
		if _, err := Parsear(entrada); err == nil {
			t.Errorf("Parsear(%q) debió fallar y no falló", entrada)
		}
	}
}
