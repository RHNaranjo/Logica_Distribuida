package mundo

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"mundos_posibles/internal/protocolo"
)

const tiempoLimite = 5 * time.Second

// Para la API dentro de cada mundo
type Controlador struct {
	repo    *Repo
	mundoID int
}

func NuevoControlador(repo *Repo, mundoID int) *Controlador {
	return &Controlador{repo: repo, mundoID: mundoID}
}

// Devolver el mux armado
func (c *Controlador) Rutas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", c.Salud)
	mux.HandleFunc("POST /configurar", c.Configurar)
	mux.HandleFunc("GET /estado", c.Estado)

	return mux
}

func (c *Controlador) Salud(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), tiempoLimite)
	defer cancel()

	configurado, err := c.repo.EstaConfigurado(ctx)
	if err != nil {
		log.Printf("[ERROR] salud: %v", err)

		responderError(w, http.StatusInternalServerError, "no se pudo consultar el estado")
		return
	}

	// Devuelve verdadero si está configurado el nodo
	responderJSON(w, http.StatusOK, protocolo.Salud{
		MundoID:     c.mundoID,
		Configurado: configurado,
	})
}

func (c *Controlador) Configurar(w http.ResponseWriter, r *http.Request) {
	var cfg protocolo.ConfiguracionMundo

	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		responderError(w, http.StatusBadRequest, "el cuerpo no es un JSON válido")
		return
	}

	// Evitar que un nodo se haga pasar por otro si el coordinador se confunde
	if cfg.MundoID != c.mundoID {
		log.Printf("[AVISO] La configuración iba dirigida a otro mundo: %d", cfg.MundoID)
		responderError(w, http.StatusBadRequest, "Esta configuración es para otro mundo. Revisar los MUNDOS_URL el el coordinador")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), tiempoLimite)
	defer cancel()

	if err := c.repo.Guardar(ctx, cfg); err != nil {
		log.Printf("[ERROR] Configurar: %v", err)
		responderError(w, http.StatusInternalServerError, "No se pudo guardar la configuración")
		return
	}

	log.Printf("[INFO] Mundo %d configurado: %d verdaderas, %d falsas, %d sucesores", cfg.MundoID, len(cfg.Verdaderas), len(cfg.Falsas), len(cfg.Sucesores))

	// No hay nada que devolver
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controlador) Estado(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), tiempoLimite)
	defer cancel()

	cfg, err := c.repo.Leer(ctx)
	if errors.Is(err, ErrNoConfigurado) {
		responderError(w, http.StatusNotFound, "Este mundo todavía no se ha configurado")
		return
	}

	if err != nil {
		log.Printf("[ERROR]Estado: %v", err)
		responderError(w, http.StatusInternalServerError, "No se pudo leer el estado")
		return
	}

	responderJSON(w, http.StatusOK, cfg)
}

// Escribir cualquier valor como JSON
func responderJSON(w http.ResponseWriter, codigo int, datos any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)

	if err := json.NewEncoder(w).Encode(datos); err != nil {
		log.Printf("[ERROR] No se pudo abrir la respuesta: %v", err)
	}
}

func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responderJSON(w, codigo, protocolo.ErrorHTTP{Error: mensaje})
}
