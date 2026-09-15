package modulos

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"logica/internal/historial"
	"logica/internal/web"
)

const (
	tiempoLimite     = 10 * time.Second
	cuerpoMaximo     = 1 << 20 // un MB
	limitePorOmision = 20
	limiteMaximo     = 200
)

// Esto sí cambia entre los distintos módulos
type Resolutor func(entrada json.RawMessage) (any, error)

// Retorno si no hay lógica
var ErrNoImplementado = errors.New("[ERROR] Este módulo no ha sido implementado")

type Controlador struct {
	modulo   string
	repo     *historial.Repo
	resolver Resolutor
}

func NuevoControlador(modulo string, repo *historial.Repo, resolver Resolutor) *Controlador {
	return &Controlador{modulo: modulo, repo: repo, resolver: resolver}
}

func (c *Controlador) Rutas() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /salud", c.Salud)
	mux.HandleFunc("POST /resolver", c.Resolver)
	mux.HandleFunc("GET /historial", c.Historial)

	return mux
}

func (c *Controlador) Salud(w http.ResponseWriter, r *http.Request) {
	web.JSON(w, http.StatusOK, map[string]any{
		"modulo":    c.modulo,
		"instancia": w.Header().Get("X-Nodo"),
		"ok":        true,
	})
}

func (c *Controlador) Resolver(w http.ResponseWriter, r *http.Request) {
	instancia := w.Header().Get("X-Nodo")

	// Tope a la memoria para no morirnos
	r.Body = http.MaxBytesReader(w, r.Body, cuerpoMaximo)

	entrada, err := io.ReadAll(r.Body)
	if err != nil {
		web.Fallo(w, http.StatusBadRequest, "no se pudo leer el cuerpo de la petición")
		return
	}

	if !json.Valid(entrada) {
		web.Fallo(w, http.StatusBadRequest, "el cuerpo no es un JSON válido")
		return
	}

	resultado, errResolver := c.resolver(entrada)

	// Anotar intentos que funcionaron y que no
	c.anotar(r.Context(), instancia, entrada, resultado, errResolver)

	if errResolver != nil {
		codigo := http.StatusBadRequest

		if errors.Is(errResolver, ErrNoImplementado) {
			codigo = http.StatusNotImplemented
		}

		web.Fallo(w, codigo, errResolver.Error())
		return
	}

	// Devolver el resultado
	web.JSON(w, http.StatusOK, map[string]any{
		"instancia": instancia,
		"resultado": resultado,
	})
}

func (c *Controlador) Historial(w http.ResponseWriter, r *http.Request) {
	limite := limitePorOmision

	if crudo := r.URL.Query().Get("limite"); crudo != "" {
		n, err := strconv.Atoi(crudo)

		if err != nil || n <= 0 || n > limiteMaximo {
			web.Fallo(w, http.StatusBadRequest, "el valor del entero debe estar entre 1 y 200")
			return
		}

		// Reemplazar por límite nuevo
		limite = n
	}

	// Contexto con tiempo válido
	ctx, cancel := context.WithTimeout(r.Context(), tiempoLimite)
	defer cancel()

	// Registros recientes
	registros, err := c.repo.Listar(ctx, c.modulo, limite)
	if err != nil {
		log.Printf("[ERROR] historial de %s: %v", c.modulo, err)

		// Ocultar detalles
		web.Fallo(w, http.StatusInternalServerError, "no se pudo leer el historial")
		return
	}

	// Devolver los registros
	web.JSON(w, http.StatusOK, registros)
}

// Guardar la base
func (c *Controlador) anotar(ctx context.Context, instancia string, entrada []byte, resultado any, errResolver error) {
	// Crear el registro
	reg := historial.Registro{
		Modulo:    c.modulo,
		Instancia: instancia,
		Entrada:   entrada,
	}

	// Guardar el error por si acaso
	if errResolver != nil {
		reg.Error = errResolver.Error()
	} else if resultado != nil {
		// Conservar historial
		crudo, err := json.Marshal(resultado)

		if err != nil {
			log.Printf("[AVISO] No se pudo serializar el resultado de la instancia %s: %v", instancia, err)
		} else {
			// Guardar por so acaso
			reg.Resultado = crudo
		}
	}

	// Limitar el tiempo
	ctxAnotar, cancel := context.WithTimeout(ctx, tiempoLimite)
	defer cancel()

	if err := c.repo.Guardar(ctxAnotar, reg); err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
