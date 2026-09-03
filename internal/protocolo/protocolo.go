package protocolo

// Define los mensajes que viajan por la red del coordinador

// Mundo alcanzable junto con la dirección de su proceso
type Sucesor struct {
	MundoID int    `json:"mundo_id"`
	URL     string `json:"url"`
}

// Todo lo que un mundo sabe de sí mismo
type ConfiguracionMundo struct {
	MundoID    int       `json:"mundo_id"`
	Verdaderas []string  `json:"verdaderas"`
	Falsas     []string  `json:"falsas"`
	Sucesores  []Sucesor `json:"sucesores"`
}

// Respuesta de GET /salud
type Salud struct {
	MundoID     int  `json:"mundo_id`
	Configurado bool `json:"configurado"`
}

type ErrorHTTP struct {
	Error string `json:"error"`
}
