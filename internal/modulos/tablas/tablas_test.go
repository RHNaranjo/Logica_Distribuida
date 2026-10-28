package tablas

import (
	"encoding/json"
	"slices"
	"testing"

	"logica/internal/formula"
)

func TestResolver(t *testing.T) {
	casos := []struct {
		entrada       string
		renglones     int
		clasificacion string
	}{
		{`{"formula": "p | ~p"}`, 2, "tautologia"},
		{`{"formula": "p & ~p"}`, 2, "contradiccion"},
		{`{"formula": "p -> q"}`, 4, "contingencia"},
		{`{"formula": "p & (q -> r)"}`, 8, "contingencia"},
	}

	for _, c := range casos {
		salida, err := Resolver(json.RawMessage(c.entrada))
		if err != nil {
			t.Fatalf("%s: error inesperado: %v", c.entrada, err)
		}

		tabla := salida.(formula.Tabla)

		if len(tabla.Renglones) != c.renglones {
			t.Errorf("%s: se esperaban %d renglones, hubo %d", c.entrada, c.renglones, len(tabla.Renglones))
		}

		if tabla.Clasificacion != c.clasificacion {
			t.Errorf("%s: se esperaba %s, hubo %s", c.entrada, c.clasificacion, tabla.Clasificacion)
		}
	}
}

func TestResolverErrores(t *testing.T) {
	casos := []string{
		`no es json`,
		`{"formula": ""}`,
		`{"formula": "p &"}`,
		`{"formula": "[]p"}`,
		`{"formula": "a & b & c & d & e & f & g & h & i & j & k"}`,
	}

	for _, entrada := range casos {
		if _, err := Resolver(json.RawMessage(entrada)); err == nil {
			t.Errorf("%s debió fallar y no falló", entrada)
		}
	}
}

// Las columnas van de adentro hacia afuera, sin el paréntesis exterior
func TestColumnas(t *testing.T) {
	salida, err := Resolver(json.RawMessage(`{"formula": "p & (q -> r)"}`))
	if err != nil {
		t.Fatal(err)
	}

	tabla := salida.(formula.Tabla)
	esperadas := []string{"q -> r", "p & (q -> r)"}

	if !slices.Equal(tabla.Columnas, esperadas) {
		t.Errorf("columnas %v, se esperaban %v", tabla.Columnas, esperadas)
	}

	// Renglón V F F: q -> r es V y p & (q -> r) es V
	if r := tabla.Renglones[4]; !slices.Equal(r.Parciales, []bool{true, true}) {
		t.Errorf("renglón VFF: parciales %v, se esperaba [true true]", r.Parciales)
	}
}
