package basededatos

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Conexión a Postgres, cada mundo se conecta con la suya

// Conectar a Postgres
func Abrir(dsn string, intentos int, espera time.Duration) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] No se pudo abrir la conexión: %w", err)
	}

	if err := ping(db, intentos, espera); err != nil {
		// Si falla la conexión, se cierra el pool para que no haya recursos colgados
		db.Close()
		return nil, err
	}

	return db, nil
}

// La función de Ping es la que genera la conexión. Se puede incrementar el margen de tiempo si no es suficiente
func ping(db *sql.DB, intentos int, espera time.Duration) error {
	var err error

	for i := 1; i <= intentos; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)

		cancel()

		if err == nil {
			return nil
		}

		log.Printf("[AVISO] Intento %d/%d de conexión fallido: %v", i, intentos, err)

		if i < intentos {
			time.Sleep(espera)
		}
	}

	return fmt.Errorf("Agotados %d intentos de conexión: %w", intentos, err)
}
