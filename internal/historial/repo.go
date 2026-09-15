package historial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Fila del historial
type Registro struct {
	ID        int64           `json:"id"`
	Modulo    string          `json:"modulo"`
	Instancia string          `json:"instancia"`
	Entrada   json.RawMessage `json:"entrada"`
	Resultado json.RawMessage `json:"resultado.omitempty"`
	Error     string          `json:"error,omitempty"`
	CreadoEn  time.Time       `json:"creado_en"`
}

// Conexión a DB
type Repo struct {
	db *sql.DB
}

func NuevoRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// Anota el intento
func (r *Repo) Guardar(ctx context.Context, reg Registro) error {
	const consulta = `
		INSERT INTO historial (modulo, instancia, entrada, resultado, error)
		VALUES ($1, $2, $3, $4, $5)
		`

	// nil -> NULL
	var resultado any
	if len(reg.Resultado) > 0 {
		resultado = string(reg.Resultado)
	}

	var errorTexto any
	if reg.Error != "" {
		errorTexto = reg.Error
	}

	// Ejecutar la consulta SQL
	_, err := r.db.ExecContext(ctx, consulta, reg.Modulo, reg.Instancia, string(reg.Entrada), resultado, errorTexto)
	if err != nil {
		return fmt.Errorf("[ERROR] No se pudo guardar en el historial: %w", err)
	}

	return nil
}

// Devuelve últimos registros de un nodo
func (r *Repo) Listar(ctx context.Context, modulo string, limite int) ([]Registro, error) {
	const consulta = `
		SELECT id, modulo, instancia, entrada, resultado, error, creado_en 
		FROM historial 
		WHERE modulo = $1 
		ORDER BY creado_en DESC, id DESC 
		LIMIT $2
		`
	// Obtener las filas de la consulta
	filas, err := r.db.QueryContext(ctx, consulta, modulo, limite)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] No se pudo consultar el historial: %w", err)
	}
	defer filas.Close()

	registros := make([]Registro, 0, limite)

	for filas.Next() {
		var reg Registro
		var entrada, resultado []byte
		var errorTexto sql.NullString

		if err := filas.Scan(
			&reg.ID, &reg.Modulo, &reg.Instancia,
			&entrada, &resultado, &errorTexto, &reg.CreadoEn,
		); err != nil {
			return nil, fmt.Errorf("[ERROR] No se pudo leer un registro: %w", err)
		}

		reg.Entrada = entrada
		reg.Resultado = resultado
		reg.Error = errorTexto.String

		registros = append(registros, reg)
	}

	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("[ERROR] Hubo un problema al recorrer el historial: %w", err)
	}

	return registros, nil
}
