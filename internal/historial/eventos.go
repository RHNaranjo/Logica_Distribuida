package historial

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"logica/internal/kafkax"
)

// Historial del módulo
func ConectarKafka(broker, modulo string, repo *Repo) (*kafka.Writer, error) {
	topico := "historial." + modulo
	grupo := "historial-" + modulo

	// Crear el tópico con una partición por instancia
	if err := kafkax.AsegurarTopico(broker, topico, 3); err != nil {
		return nil, err
	}

	// Guardar en postgres lo que haya en el tópico con una gorutina
	go func() {
		for {
			err := kafkax.Consumir(context.Background(), broker, topico, grupo, repo.GuardarEvento)
			log.Printf("[ERROR] El consumidor de %s se detuvo: %v. Reiniciando", topico, err)
			time.Sleep(5 * time.Second)
		}
	}()

	log.Printf("[INFO] Historial de %s conectado a Kafka (%s)", modulo, topico)

	return kafkax.NuevoEscritor(broker, topico), nil
}

// Mandar el registro al tópico
func Publicar(ctx context.Context, escritor *kafka.Writer, reg Registro) error {
	// Identificar el evento
	reg.EventoID = uuid.NewString()
	reg.CreadoEn = time.Now().UTC()

	valor, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	// La llave es la instancia
	return escritor.WriteMessages(ctx, kafka.Message{Key: []byte(reg.Instancia), Value: valor})
}

// Guardar la base en un evento que llegó del tópico
func (r *Repo) GuardarEvento(ctx context.Context, msg kafka.Message) error {
	var reg Registro

	if err := json.Unmarshal(msg.Value, &reg); err != nil {
		// Brincar mensajes rotos
		log.Printf("[ERROR] Evento ilegible en el offset %d: %v", msg.Offset, err)
		return nil
	}

	return r.Guardar(ctx, reg)
}
