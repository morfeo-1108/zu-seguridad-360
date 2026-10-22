// Package bus wraps RabbitMQ for the ZU services: a consumer that survives
// broker restarts and retries failed messages through a TTL queue before
// dead-lettering them, and a publisher that reconnects on demand.
//
// Topology per consumer queue Q:
//
//	Q        durable, bound to the given exchange/keys, DLX → Q.retry
//	Q.retry  TTL RetryDelay, DLX → Q (redelivery)
//	Q.dead   messages that failed MaxAttempts times or permanently
package bus

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Permanent wraps errors that retrying cannot fix (malformed payloads).
type permanent struct{ error }

func Permanent(err error) error { return permanent{err} }

type Binding struct{ Exchange, Key string }

type Consumer struct {
	URL         string
	Queue       string
	Bindings    []Binding
	Handler     func(context.Context, amqp.Delivery) error
	MaxAttempts int
	RetryDelay  time.Duration

	up atomic.Bool
}

func (c *Consumer) Up() bool { return c.up.Load() }

func (c *Consumer) Run(ctx context.Context) {
	if c.MaxAttempts == 0 {
		c.MaxAttempts = 5
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = 30 * time.Second
	}
	backoff := time.Second
	for ctx.Err() == nil {
		err := c.runOnce(ctx)
		c.up.Store(false)
		if ctx.Err() != nil {
			return
		}
		log.Printf("consumer %s disconnected: %v (retry in %s)", c.Queue, err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (c *Consumer) runOnce(ctx context.Context) error {
	conn, err := amqp.Dial(c.URL)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	retry, dead := c.Queue+".retry", c.Queue+".dead"
	if _, err := ch.QueueDeclare(c.Queue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": "", "x-dead-letter-routing-key": retry,
	}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(retry, true, false, false, false, amqp.Table{
		"x-message-ttl": int32(c.RetryDelay / time.Millisecond), "x-dead-letter-exchange": "", "x-dead-letter-routing-key": c.Queue,
	}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(dead, true, false, false, false, nil); err != nil {
		return err
	}
	for _, b := range c.Bindings {
		// Passive: exchanges belong to their producers (Core declares zu.*); a
		// consumer only checks they exist, which needs no configure permission.
		if err := ch.ExchangeDeclarePassive(b.Exchange, "topic", true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(c.Queue, b.Key, b.Exchange, false, nil); err != nil {
			return err
		}
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(c.Queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	c.up.Store(true)
	log.Printf("consumer attached to %s", c.Queue)
	closed := conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-closed:
			return e
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			c.handle(ctx, ch, d, dead)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, ch *amqp.Channel, d amqp.Delivery, dead string) {
	err := c.Handler(ctx, d)
	if err == nil {
		d.Ack(false)
		return
	}
	var p permanent
	n := Attempts(d, c.Queue)
	if errors.As(err, &p) || n >= c.MaxAttempts-1 {
		log.Printf("%s: dead-lettering %s after %d attempt(s): %v", c.Queue, d.RoutingKey, n+1, err)
		if perr := ch.PublishWithContext(ctx, "", dead, false, false, amqp.Publishing{
			ContentType: d.ContentType, DeliveryMode: amqp.Persistent, Body: d.Body,
			Headers: amqp.Table{"x-zu-reason": err.Error(), "x-zu-routing-key": d.RoutingKey},
		}); perr != nil {
			d.Nack(false, true)
			return
		}
		d.Ack(false)
		return
	}
	log.Printf("%s: %s failed (attempt %d/%d): %v", c.Queue, d.RoutingKey, n+1, c.MaxAttempts, err)
	d.Nack(false, false)
}

// Attempts counts how many times the message was dead-lettered out of queue.
func Attempts(d amqp.Delivery, queue string) int {
	deaths, _ := d.Headers["x-death"].([]interface{})
	for _, x := range deaths {
		if t, ok := x.(amqp.Table); ok && t["queue"] == queue {
			if n, ok := t["count"].(int64); ok {
				return int(n)
			}
		}
	}
	return 0
}

// Publisher keeps one channel open and reconnects on the next publish after
// a failure. Messages are persistent; publishing errors are returned, never
// swallowed, so callers can record them.
type Publisher struct {
	URL string

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (p *Publisher) Publish(ctx context.Context, exchange, key string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for attempt := 0; attempt < 2; attempt++ {
		if p.ch == nil || p.ch.IsClosed() {
			if err := p.connect(); err != nil {
				return err
			}
		}
		err := p.ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{
			ContentType: "application/json", DeliveryMode: amqp.Persistent, Timestamp: time.Now(), Body: body,
		})
		if err == nil {
			return nil
		}
		p.ch = nil
		if attempt == 1 {
			return err
		}
	}
	return nil
}

func (p *Publisher) connect() error {
	if p.conn != nil && !p.conn.IsClosed() {
		p.conn.Close()
	}
	conn, err := amqp.Dial(p.URL)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	p.conn, p.ch = conn, ch
	return nil
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		p.conn.Close()
	}
}
