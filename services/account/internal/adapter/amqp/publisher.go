// Package amqp publishes Account's events to RabbitMQ.
package amqp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/neennera/fishertimer/pkg/events"
)

const (
	// confirmTimeout bounds how long one publish waits for the broker's ack.
	confirmTimeout = 5 * time.Second
	// queueSize is how many profile changes wait for the broker at most.
	queueSize = 256

	minBackoff = time.Second
	maxBackoff = 30 * time.Second
)

// Publisher sends profile.updated to the fisher.account exchange with
// publisher confirms.
//
// PublishProfileUpdated only queues the event in memory and returns at once; a
// background worker sends it, in order, and keeps retrying with backoff while
// RabbitMQ is down. The connection is opened lazily, so the service starts and
// serves renames with no broker.
//
// Limit: events still waiting when the service stops or crashes are lost (no
// outbox table here, unlike Study Session). A lost event leaves the old name in
// rooms and on the leaderboard until the next rename.
type Publisher struct {
	url   string
	queue chan events.ProfileUpdated

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel

	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func NewPublisher(url string) *Publisher {
	p := &Publisher{
		url:   url,
		queue: make(chan events.ProfileUpdated, queueSize),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go p.run()
	return p
}

// PublishProfileUpdated queues the change and returns immediately.
func (p *Publisher) PublishProfileUpdated(userID, displayName string, at time.Time) {
	ev := events.ProfileUpdated{
		EventID:     newEventID(),
		OccurredAt:  at.UTC(),
		UserID:      userID,
		DisplayName: displayName,
	}
	select {
	case p.queue <- ev:
	default:
		// The broker has been unreachable long enough to fill the queue.
		log.Printf("account: profile event queue full, dropping profile.updated for user %s", userID)
	}
}

// run sends queued events one at a time, retrying each until the broker takes it.
func (p *Publisher) run() {
	defer close(p.done)
	for {
		select {
		case <-p.stop:
			return
		case ev := <-p.queue:
			p.sendWithRetry(ev)
		}
	}
}

func (p *Publisher) sendWithRetry(ev events.ProfileUpdated) {
	backoff := minBackoff
	for {
		err := p.send(ev)
		if err == nil {
			return
		}
		log.Printf("account: could not publish profile.updated for user %s (%v); retrying in %s", ev.UserID, err, backoff)
		select {
		case <-p.stop:
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (p *Publisher) send(ev events.ProfileUpdated) error {
	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	ch, err := p.channel()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), confirmTimeout)
	defer cancel()
	confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx,
		events.AccountExchange,
		events.RoutingKeyProfileUpdated,
		false, // mandatory: the topology guarantees the queues exist
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // survives a broker restart
			MessageId:    ev.EventID,
			Timestamp:    ev.OccurredAt,
			Type:         events.RoutingKeyProfileUpdated,
			AppId:        "account",
			Body:         body,
		},
	)
	if err != nil {
		p.reset()
		return fmt.Errorf("publish: %w", err)
	}
	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		p.reset()
		return fmt.Errorf("confirm: %w", err)
	}
	if !acked {
		return errors.New("broker nacked the event")
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
	log.Printf("account: connected to RabbitMQ, publishing to exchange %s", events.AccountExchange)
	return ch, nil
}

func (p *Publisher) reset() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch = nil, nil
}

// Close stops the worker and releases the connection. Events still queued are dropped.
func (p *Publisher) Close() {
	p.once.Do(func() { close(p.stop) })
	<-p.done
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reset()
}

func newEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}
