package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
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

// Devolver direcciones de los nodo-mundo
// urls[0] es el mundo0, urls[1] es el mundo1
func MundosURLs() ([]string, error) {
	valor := os.Getenv("MUNDOS_URLS")

	if valor == "" {
		return nil, fmt.Errorf("[ERROR] Falta la variable de entorno MUNDOS_URLS")
	}

	partes := strings.Split(valor, ",")

	// Crear el espacio en memoria para los URLs
	urls := make([]string, 0, len(partes))

	for _, parte := range partes {
		parte = strings.TrimSpace(parte)
		if parte == "" {
			continue
		}

		// Quitar la diagnoal
		// Objetivo: http://mundo0:8080//configurar
		parte = strings.TrimSuffix(parte, "/")

		u, err := url.Parse(parte)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("[ERROR] La dirección %q no es válida (¿le falta el http://?)", parte)
		}

		urls = append(urls, parte)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("[ERROR] MUNDOS_URLS no contiene ninguna dirección")
	}

	return urls, nil
}

// Determinar cuántas instancias levanta el proceso
// Cada una es una gorutina con su propio HTTP y puerto
func NumInstancias() int {
	return getEnvInt("INSTANCIAS", 3)
}

// Puerto de la primera instancia; las demás instancias se van sumando
func PuertoBase(respaldo int) int {
	return getEnvInt("PUERTO_BASE", respaldo)
}

// Nombre con el que las instancias se anuncian ante el LB
func HostAnunciado(respaldo string) string {
	return getEnv("HOST_ANUNCIO", respaldo)
}

// Dirección del POST /registrar
func URLRegistro() string {
	return os.Getenv("REGISTRO_URL")
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
