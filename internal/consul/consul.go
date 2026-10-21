package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Instancia registrada
type Instancia struct {
	Nombre string
	Sana   bool
}

// Datos con los que una instancia se da de alta
type Servicio struct {
	ID     string // tablas-0
	Nombre string // tablas
	Host   string // nombre para Docker
	Puerto int    // 9021
}

// Cliente con tiempo límite
var cliente = &http.Client{Timeout: 3 * time.Second}

// Dar de alta la instancia
func Registrar(direccion string, s Servicio) {
	cuerpo, err := json.Marshal(map[string]any{
		"ID":      s.ID,
		"Name":    s.Nombre,
		"Address": s.Host,
		"Port":    s.Puerto,
		"Check": map[string]any{
			"GRPC":                           fmt.Sprintf("%s:%d", s.Host, s.Puerto),
			"GRPCUseTLS":                     false,
			"Interval":                       "5s",
			"Timeout":                        "2s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	})
	if err != nil {
		log.Printf("[ERROR] No se pudo armar el registro de %s: %v", s.ID, err)
		return
	}

	url := "http://" + direccion + "/v1/agent/service/register"

	for intento := 1; intento <= 10; intento++ {
		err = enviar(http.MethodPut, url, cuerpo)

		if err == nil {
			log.Printf("[INFO] %s registrada en Consul (%s)", s.ID, direccion)
			return
		}

		log.Printf("[AVISO] Intento %d/10 de registrar %s: %v", intento, s.ID, err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("[ERROR] %s no pudo registrarse; nginx no la verá", s.ID)
}

// Quitar la instancia para que nginx la deje en paz
func Borrar(direccion, id string) {
	url := "http://" + direccion + "/v1/agent/service/deregister/" + id

	if err := enviar(http.MethodPut, url, nil); err != nil {
		log.Printf("[AVISO] No se pudo borrar %s de Consul: %v", id, err)
		return
	}

	log.Printf("[INFO] %s borrada de Consul", id)
}

// Leer todas las instancias y su salud por módulo
func Estado(direccion string) (map[string][]Instancia, error) {
	resp, err := cliente.Get("http://" + direccion + "/v1/health/state/any")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Consul respondió %d", resp.StatusCode)
	}

	var checks []struct {
		ServiceID   string
		ServiceName string
		Status      string
	}
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		return nil, err
	}

	estado := make(map[string][]Instancia)

	for _, c := range checks {
		// El check del nodo de Consul viene vacío
		if c.ServiceID == "" {
			continue
		}

		estado[c.ServiceName] = append(estado[c.ServiceName], Instancia{
			Nombre: c.ServiceID,
			Sana:   c.Status == "passing",
		})
	}

	return estado, nil
}

// Mandar una petición a Consul
func enviar(metodo, url string, cuerpo []byte) error {
	req, err := http.NewRequest(metodo, url, bytes.NewReader(cuerpo))
	if err != nil {
		return err
	}

	resp, err := cliente.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Consul respondió %d", resp.StatusCode)
	}

	return nil
}
