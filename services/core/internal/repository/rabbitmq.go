package repository

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/core/config"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(cfg *config.Config) *RabbitMQ {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open RabbitMQ channel: %v", err)
	}

	// Declare exchanges
	exchanges := []string{"zu.companies", "zu.employees", "zu.notifications", "zu.documents"}
	for _, ex := range exchanges {
		ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil)
	}

	log.Println("✅ RabbitMQ connected")
	return &RabbitMQ{Conn: conn, Channel: ch}
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return err
	}
	return r.Conn.Close()
}

func (r *RabbitMQ) Publish(exchange, routingKey string, body []byte) error {
	return r.Channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}
