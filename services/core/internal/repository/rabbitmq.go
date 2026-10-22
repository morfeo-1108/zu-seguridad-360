package repository

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/core/config"
)

var coreExchanges = []string{"zu.companies", "zu.employees", "zu.notifications", "zu.documents"}

// RabbitMQ publishes Core's domain events. The first connection is required
// at startup (as before); after that a dropped connection is re-dialed on the
// next publish instead of leaving Core publishing into a dead channel.
type RabbitMQ struct {
	url  string
	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQ(cfg *config.Config) *RabbitMQ {
	r := &RabbitMQ{url: cfg.RabbitMQURL}
	if err := r.connect(); err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	log.Println("✅ RabbitMQ connected")
	return r
}

func (r *RabbitMQ) connect() error {
	if r.conn != nil && !r.conn.IsClosed() {
		r.conn.Close()
	}
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	for _, ex := range coreExchanges {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			conn.Close()
			return err
		}
	}
	r.conn, r.ch = conn, ch
	return nil
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

// Publish sends a persistent message, reconnecting once if the channel was
// lost. Errors are returned and logged, never swallowed.
func (r *RabbitMQ) Publish(exchange, routingKey string, body []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		if r.ch == nil || r.ch.IsClosed() {
			if err = r.connect(); err != nil {
				continue
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = r.ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
			ContentType: "application/json", DeliveryMode: amqp.Persistent, Timestamp: time.Now(), Body: body,
		})
		cancel()
		if err == nil {
			return nil
		}
		r.ch = nil
	}
	log.Printf("rabbitmq publish %s/%s failed: %v", exchange, routingKey, err)
	return err
}
