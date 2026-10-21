package kafkax

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// Crea el tópico y reintenta si Kafka tarda en arrancar
func AsegurarTopico(broker, topico string, particiones int) error {
	var err error

	for intento := 1; intento <= 15; intento++ {
		err = crearTopico(broker, topico, particiones)
		if err == nil {
			return nil
		}

		log.Printf("[AVISO] Intento %d/15 de crear el tópico %s: %v", intento, topico, err)
		time.Sleep(2 * time.Second)
	}

	return err
}

// Para que Kafka cree un nuevo tópico (si no existe ya)
func crearTopico(broker, topico string, particiones int) error {
	conexion, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conexion.Close()

	// Sólo el controlador del cluster puede crear tópicos
	controlador, err := conexion.Controller()
	if err != nil {
		return err
	}

	ctrl, err := kafka.Dial("tcp", net.JoinHostPort(controlador.Host, strconv.Itoa(controlador.Port)))
	if err != nil {
		return err
	}
	defer ctrl.Close()

	return ctrl.CreateTopics(kafka.TopicConfig{
		Topic:             topico,
		NumPartitions:     particiones,
		ReplicationFactor: 1, // sólo un nodo de kafka
	})
}

// Crear el productor
func NuevoEscritor(broker, topico string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topico,
		Balancer:     &kafka.Hash{},    // partición = hash(llave)
		RequiredAcks: kafka.RequireAll, // Esperar a que kafka lo guarde
		BatchTimeout: 10 * time.Millisecond,
	}
}

// Procesar un mensaje
type Manejador func(ctx context.Context, msg kafka.Message) error

// Unirse al grupo y procesar mensajes hasta que se cancele el ctx
// Offset se confirma después de procesar (At Least Once)
func Consumir(ctx context.Context, broker, topico, grupo string, manejar Manejador) error {
	lector := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       topico,
		GroupID:     grupo,
		StartOffset: kafka.FirstOffset, // Un grupo nuevo comienza desde el inicio
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer lector.Close()

	for {
		// Leer el siguiente mensaje sin confirmarlo
		msg, err := lector.FetchMessage(ctx)
		if err != nil {
			return err
		}

		// Reintentar el mismo mensaje hasta procesarlo
		for {
			err := manejar(ctx, msg)
			if err == nil {
				break
			}

			log.Printf("[AVISO] %s partición %d offset %d: %v. Reintentar", topico, msg.Partition, msg.Offset, err)
			time.Sleep(2 * time.Second)
		}

		if err := lector.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}
