package loadbalancer

import (
	"cmp"
	"log"
	"slices"
	"sync"
)

// lO que sabe de una instancia de un modulo
type Instancia struct {
	Nombre    string `json:"instancia"`
	URL       string `json:"url"`
	Sana      bool   `json:"sana"`
	Atendidas int    `json:"atendidas"`
}

// Vigilante que revisa instancia
type objetivo struct {
	modulo string
	nombre string
	url    string
}

// Directorio de servicios
type Registro struct {
	mu      sync.Mutex
	modulos map[string][]Instancia // "tablas" -> [tablas-0, tablas-1, tablas-2, ...]

	siguiente map[string]int // a quién le toca el round-robin
}

func NuevoRegistro() *Registro {
	return &Registro{
		modulos:   make(map[string][]Instancia),
		siguiente: make(map[string]int),
	}
}

// Dar de alta una instancia o revivirla
func (r *Registro) Agregar(modulo, nombre, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lista := r.modulos[modulo]

	for i := range lista {
		if lista[i].Nombre == nombre {
			// Ya se conoce
			lista[i].URL = url
			lista[i].Sana = true
			log.Printf("[INFO] %s se volvió a registrar en %s", nombre, url)
			return
		}
	}

	lista = append(lista, Instancia{Nombre: nombre, URL: url, Sana: true})

	// Ordenar por nombre
	slices.SortFunc(lista, func(a, b Instancia) int {
		return cmp.Compare(a.Nombre, b.Nombre)
	})

	r.modulos[modulo] = lista
	log.Printf("[INFO] Registrada %s en %s", nombre, url)
}

// Devuelve una copia de la siguiente instancia sana (round-robin)
func (r *Registro) Elegir(modulo string) (Instancia, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lista := r.modulos[modulo]
	n := len(lista)

	for i := 0; i < n; i++ {
		idx := (r.siguiente[modulo] + i) % n

		if lista[idx].Sana {
			r.siguiente[modulo] = idx + 1
			lista[idx].Atendidas++
			return lista[idx], true
		}
	}

	return Instancia{}, false
}

// Actualiza el estado de una instancia y avisa si cambió
func (r *Registro) MarcarSalud(modulo, nombre string, sana bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lista := r.modulos[modulo]

	for i := range lista {
		if lista[i].Nombre == nombre && lista[i].Sana != sana {
			lista[i].Sana = sana

			if sana {
				log.Printf("[INFO] %s volvió a responder, regresa a la rotación", nombre)
			} else {
				log.Printf("[AVISO] %s dejó de responder; sale de la rotación", nombre)
			}
		}
	}
}

// Devuelve una copia de todo el registro
func (r *Registro) Foto() map[string][]Instancia {
	r.mu.Lock()
	defer r.mu.Unlock()

	foto := make(map[string][]Instancia, len(r.modulos))

	for modulo, lista := range r.modulos {
		foto[modulo] = slices.Clone(lista)
	}

	return foto
}

// Devuelve a la instancia que va a revisarse
func (r *Registro) objetivos() []objetivo {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lista []objetivo

	for modulo, instancias := range r.modulos {
		for _, inst := range instancias {
			lista = append(lista, objetivo{modulo: modulo, nombre: inst.Nombre, url: inst.URL})
		}
	}

	return lista
}
