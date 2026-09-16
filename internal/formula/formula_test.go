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
