package config

import (
	"fmt"
	"log"
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
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		usuario, contrasena, host, puerto, nombre,
	)

	return dsn, nil
}

// Obtener el puerto en el que el proceso escucha
func PuertoHTTP(respaldo string) string {
	return ":" + getEnv("PUERTO", respaldo)
}

// Leer la variable de entorno y regresar respaldo si no se encuentra
func getEnv(clave, respaldo string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}

	return respaldo
}

// Determinar cuántas instancias levanta el proceso
// Cada una es una gorutina con su propio servidor gRPC y puerto
func NumInstancias() int {
	return getEnvInt("INSTANCIAS", 3)
}

// Puerto de la primera instancia; las demás instancias se van sumando
func PuertoBase(respaldo int) int {
	return getEnvInt("PUERTO_BASE", respaldo)
}

// Convertir getEnv pero en entero
func getEnvInt(clave string, respaldo int) int {
	valor := os.Getenv(clave)

	if valor == "" {
		return respaldo
	}

	n, err := strconv.Atoi(valor)
	if err != nil || n <= 0 {
		log.Printf("[AVISO] %s=%q no es un entero positivo; se usa %d", clave, valor, respaldo)
		return respaldo
	}

	return n
}

// Dirección gRPC a la que el middlware manda sus cosas
func DestinoGRPC(respaldo string) string {
	return getEnv("GRPC_DESTINO", respaldo)
}

// Carpeta con el frontend que sirve el contenedor
func DirectorioWeb(respaldo string) string {
	return getEnv("DIRECTORIO_WEB", respaldo)
}
