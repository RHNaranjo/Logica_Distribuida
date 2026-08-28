package config

import (
	"fmt"
	"os"
	"strconv"
)

// Cada nodo apunta a su base de datos. DSN arma conexión a Postgres de su propio nodo
func DSN() (string, error) {
	host := getEnv("DB_HOST", "localhost")
	puerto := getEnv("DB_PORT", "5432")
	usuario := os.Getenv("DB_USER")
	contrasena := os.Getenv("DB_PASSWORD")
	nombre := os.Getenv("DB_NAME")

	// Validaciones
	if usuario == "" {
		return "", fmt.Errorf("Falta la variable de entorno DB_USER")
	}
	if contrasena == "" {
		return "", fmt.Errorf("Falta la variable de entorno DB_PASSWORD")
	}
	if nombre == "" {
		return "", fmt.Errorf("Falta la variable de entorno DB_NAME")
	}

	// Sprintf sólo admite valores alfanumericos, por lo que debería validarse eso en la contraseña (pendiente para otro día)
	dsn := fmt.Sprintf(
		"postgres://%s : %s @ %s : %s / %s?sslmode=disable",
		usuario, contrasena, host, puerto, nombre,
	)

	return dsn, nil
}

// Devuelve el identificador del mundo del proceso
func MundoID() (int, error) {
	valor := os.Getenv("MUNDO_ID")

	if valor == "" {
		return 0, fmt.Errorf("Falta la variable de entorno MUNDO_ID")
	}

	id, err := strconv.Atoi(valor)
	if err != nil {
		// Usé %w para incluir el contexto del error, que se pierde si hubiera utilizado %v
		return 0, fmt.Errorf("MUNDO_ID debe ser un número entero, pero se recibió %q: %w", valor, err)
	}

	if id < 0 {
		return 0, fmt.Errorf("MUNDO_ID debe ser mayo o igual a 0, se recibió: %d", id)
	}

	return id, nil
}

// Obtener el puerto en el que el proceso escucha
func PuertoHTTP() string {
	return ":" + getEnv("PUERTO", "8080")
}

// Leer la variable de entorno y regresar respaldo si no se encuentra
func getEnv(clave, respaldo string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}

	return respaldo
}
