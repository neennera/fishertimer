package amqp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/neennera/fishertimer/pkg/events"
)

// confirmTimeout bounds how long one publish waits for the broker's ack.
const confirmTimeout = 5 * time.Second

// Publisher sends session events to the fisher.session exchange with
// publisher confirms: Publish returns nil only after RabbitMQ has taken
// responsibility for the message.
//
// The connection is opened lazily and re-opened after any failure, so the
// service starts and keeps working while RabbitMQ is down; the outbox holds
// the events until a publish succeeds again.
type Publisher struct {
	url string

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(url string) *Publisher {
	return &Publisher{url: url}
}

func (p *Publisher) Publish(ctx context.Context, routingKey, eventID string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch, err := p.channel()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()
	confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx,
		events.SessionExchange,
		routingKey,
		false, // mandatory: the topology below guarantees the queue exists
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // survives a broker restart
			MessageId:    eventID,
			Timestamp:    time.Now().UTC(),
			Type:         routingKey,
			AppId:        "study-session",
			Body:         payload,
		},
	)
	if err != nil {
		p.reset()
		return fmt.Errorf("publish %s: %w", routingKey, err)
	}
	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		p.reset()
		return fmt.Errorf("confirm %s: %w", routingKey, err)
	}
	if !acked {
		return errors.New("broker nacked " + routingKey)
	}
	return nil
}

// channel returns an open confirm-mode channel, dialing if needed.
func (p *Publisher) channel() (*amqp.Channel, error) {
	if p.ch != nil && !p.ch.IsClosed() && p.conn != nil && !p.conn.IsClosed() {
		return p.ch, nil
	}
	p.reset()

	conn, err := amqp.Dial(p.url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}
	if err := DeclareTopology(ch); err != nil {
		conn.Close()
		return nil, err
	}
	p.conn, p.ch = conn, ch
	log.Printf("study-session: connected to RabbitMQ, publishing to exchange %s", events.SessionExchange)
	return ch, nil
}

func (p *Publisher) reset() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch = nil, nil
}

// Close releases the connection.
func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reset()
}
