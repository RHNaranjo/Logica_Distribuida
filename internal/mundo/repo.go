package mundo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mundos_posibles/internal/protocolo"
)

// Cuando se pide el estado de un mundo al que el coordinador no le ha mandado nada
var ErrNoConfigurado = errors.New("Este mundo todavía no ha sido configurado")

type Repo struct {
	db *sql.DB
}

func NuevoRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// Reemplazar lo que el mundo sabe de sí mismo
func (r *Repo) Guardar(ctx context.Context, cfg protocolo.ConfiguracionMundo) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] No se pudo abrir la transacción: %w", err)
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM variables`); err != nil {
		return fmt.Errorf("[ERROR] No se pudieron borrar las variables anteriores: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sucesores`); err != nil {
		return fmt.Errorf("[ERROR] No se pudieron borrar los sucesores anteriores: %w", err)
	}

	const insVariable = `INSERT INTO variables (nombre, valor) VALUES ($1, $2)`

	for _, nombre := range cfg.Verdaderas {
		if _, err := tx.ExecContext(ctx, insVariable, nombre, true); err != nil {
			return fmt.Errorf("[ERROR] No se pudo guardar la variable %q: %w", nombre, err)
		}
	}

	for _, nombre := range cfg.Falsas {
		if _, err := tx.ExecContext(ctx, insVariable, nombre, false); err != nil {
			return fmt.Errorf("[ERROR] No se pudo guardar la variable ~%q: %w", nombre, err)
		}
	}

	const insSucesor = `INSERT INTO sucesores (mundo_id, url) VALUES ($1, $2)`

	for _, s := range cfg.Sucesores {
		if _, err := tx.ExecContext(ctx, insSucesor, s.MundoID, s.URL); err != nil {
			return fmt.Errorf("[ERROR] No se pudo guardar el sucesor %d: %w", s.MundoID, err)
		}
	}

	// Insertar la fila 1 si no existe. Si ya existe, la actualiza
	const upsertConfig = `
		INSERT INTO configuracion (id, mundo_id, actualizado)
		VALUES (1, $1, now())
		ON CONFLICT (id) DO UPDATE
			SET mundo_id = EXCLUDED.mundo_id, actualizado = now()
		`

	if _, err := tx.ExecContext(ctx, upsertConfig, cfg.MundoID); err != nil {
		return fmt.Errorf("[ERROR] No se pudo marcar el mundo como configurado: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("[ERROR] No se pudo confirmar la transacción: %w", err)
	}

	return nil
}

// Responder a consultas si el mundo ya está configurado
func (r *Repo) EstaConfigurado(ctx context.Context) (bool, error) {
	var existe bool

	const query = `SELECT EXISTS (SELECT 1 FROM configuracion WHERE id = 1)`

	if err := r.db.QueryRowContext(ctx, query).Scan(&existe); err != nil {
		return false, fmt.Errorf("[ERROR] No se pudo consultar la configuración: %w", err)
	}

	return existe, nil
}

// Devuelve todo lo que el mundo sabe de sí mismo
func (r *Repo) Leer(ctx context.Context) (protocolo.ConfiguracionMundo, error) {
	var cfg protocolo.ConfiguracionMundo

	const qConfig = `SELECT mundo_id FROM configuracion WHERE id = 1`

	err := r.db.QueryRowContext(ctx, qConfig).Scan(&cfg.MundoID)

	// Analizar tipo de error
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return protocolo.ConfiguracionMundo{}, ErrNoConfigurado
	case err != nil:
		return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] No se pudo leer la configuracion: %w", err)
	}

	const qVariables = `SELECT nombre, valor FROM variables ORDER BY nombre`

	filas, err := r.db.QueryContext(ctx, qVariables)
	if err != nil {
		return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] No se pudieron leer las variables: %w", err)
	}

	defer filas.Close()

	for filas.Next() {
		var nombre string
		var valor bool

		if err := filas.Scan(&nombre, &valor); err != nil {
			return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] No se pudo leer una variable: %w", err)
		}

		// Añadir la variable a la lista
		if valor {
			cfg.Verdaderas = append(cfg.Verdaderas, nombre)
		} else {
			cfg.Falsas = append(cfg.Falsas, nombre)
		}
	}

	if err := filas.Err(); err != nil {
		return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] Error recorriendo las variables: %w", err)
	}

	const qSucesores = `SELECT mundo_id, url FROM sucesores ORDER BY mundo_id`

	filasSuc, err := r.db.QueryContext(ctx, qSucesores)
	if err != nil {
		return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] No se pudieron leer los sucesores: %w", err)
	}
	defer filasSuc.Close()

	for filasSuc.Next() {
		var s protocolo.Sucesor

		if err := filasSuc.Scan(&s.MundoID, &s.URL); err != nil {
			return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] No se pudo leer un sucesor: %w", err)
		}

		cfg.Sucesores = append(cfg.Sucesores, s)
	}

	if err := filasSuc.Err(); err != nil {
		return protocolo.ConfiguracionMundo{}, fmt.Errorf("[ERROR] No se pudo recorrer los sucesores: %w", err)
	}

	return cfg, nil
}
