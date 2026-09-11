package loadbalancer

import (
	"net/http"
	"time"
)

// Revisar cada instancia cada X tiempo
func VigilarSalud(reg *Registro, cada time.Duration) {
	cliente := &http.Client{Timeout: 2 * time.Second}

	ticker := time.NewTicker(cada)
	defer ticker.Stop()

	for range ticker.C {
		for _, obj := range reg.objetivos() {
			// Obtener la salud de los URLs uno por uno
			sana := responde(cliente, obj.url)
			reg.MarcarSalud(obj.modulo, obj.nombre, sana)
		}
	}
}

// Devuelve la salud de la instancia
func responde(cliente *http.Client, url string) bool {
	resp, err := cliente.Get(url + "/salud")
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
