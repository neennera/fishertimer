package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Reconnect backoff: starts short, doubles, capped.
const (
	minBackoff = time.Second
	maxBackoff = 30 * time.Second
)

type Consumer struct {
	url string
	uc  usecase.Usecase
}

func NewConsumer(url string, uc usecase.Usecase) *Consumer {
	return &Consumer{url: url, uc: uc}
}

// Run connects, declares the topology and consumes timer.session-events
// until ctx is cancelled. Whenever the connection drops (broker restart,
// network) it reconnects with backoff, so the Timer never silently stops
// receiving events; meanwhile they wait in the durable queue.
func (c *Consumer) Run(ctx context.Context) {
	backoff := minBackoff
	for {
		err := c.consumeOnce(ctx)
		if ctx.Err() != nil {
			log.Printf("study-timer: stopping RabbitMQ consumer")
			return
		}
		if err == nil {
			backoff = minBackoff // a healthy session ended; retry promptly
		}
		log.Printf("study-timer: RabbitMQ consumer disconnected (%v); reconnecting in %s", err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

// consumeOnce runs one connection until it closes. Returns nil if it was
// established and later lost, an error if it could not be set up.
func (c *Consumer) consumeOnce(ctx context.Context) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := DeclareTopology(ch); err != nil {
		return err
	}
	// Prefetch 10 for fair dispatch.
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(events.TimerSessionQueue, "study-timer-consumer",
		false, // manual ack
		false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("study-timer: RabbitMQ consumer worker started on queue %s", events.TimerSessionQueue)

	closed := conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return nil
		case amqpErr := <-closed:
			if amqpErr != nil {
				return nil
			}
			return errors.New("connection closed")
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			c.handleDelivery(ctx, msg)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, msg amqp.Delivery) {
	switch msg.RoutingKey {
	case events.RoutingKeyJoined:
		var ev events.ParticipantJoined
		if !decode(msg, &ev) {
			return
		}
		if ev.EventID == "" || ev.SessionID == "" || ev.UserID == "" {
			reject(msg, "malformed payload")
			return
		}
		log.Printf("[Consumer] Participant joined: event_id=%s, session_id=%s, user_id=%s", ev.EventID, ev.SessionID, ev.UserID)
		settle(msg, c.uc.OpenParticipantTimer(ctx, ev.SessionID, ev.UserID, ev.EventID))

	case events.RoutingKeyLeft:
		var ev events.ParticipantLeft
		if !decode(msg, &ev) {
			return
		}
		if ev.EventID == "" || ev.SessionID == "" || ev.UserID == "" {
			reject(msg, "malformed payload")
			return
		}
		log.Printf("[Consumer] Processing participant left: event_id=%s, session_id=%s, user_id=%s, reason=%s",
			ev.EventID, ev.SessionID, ev.UserID, ev.Reason)
		settle(msg, c.uc.FinalizeParticipantTimer(ctx, ev.SessionID, ev.UserID, ev.EventID, ev.Reason))

	case events.RoutingKeyEnded:
		var ev events.SessionEnded
		if !decode(msg, &ev) {
			return
		}
		if ev.EventID == "" || ev.SessionID == "" {
			reject(msg, "malformed payload")
			return
		}
		log.Printf("[Consumer] Processing session ended: event_id=%s, session_id=%s, reason=%s",
			ev.EventID, ev.SessionID, ev.Reason)
		settle(msg, c.uc.FinalizeSessionTimers(ctx, ev.SessionID, ev.EventID, ev.Reason))

	default:
		log.Printf("[Consumer] Unknown routing key %s. Acknowledging and skipping", msg.RoutingKey)
		_ = msg.Ack(false)
	}
}

// decode parses the payload; an unparsable message goes to the DLQ.
func decode(msg amqp.Delivery, v any) bool {
	if err := json.Unmarshal(msg.Body, v); err != nil {
		reject(msg, err.Error())
		return false
	}
	return true
}

// reject sends a message that can never succeed to the DLQ.
func reject(msg amqp.Delivery, why string) {
	log.Printf("[Consumer] %s message rejected (%s). Nacking to DLQ", msg.RoutingKey, why)
	_ = msg.Nack(false, false)
}

// settle acks a handled event, sends one with invalid ids to the DLQ, or
// requeues it after a transient failure (e.g. timer_db down) so it is
// processed again later.
func settle(msg amqp.Delivery, err error) {
	if errors.Is(err, domain.ErrInvalid) {
		reject(msg, err.Error())
		return
	}
	if err != nil {
		log.Printf("[Consumer] Transient error on %s: %v. Requeueing", msg.RoutingKey, err)
		_ = msg.Nack(false, true)
		return
	}
	_ = msg.Ack(false)
	log.Printf("[Consumer] %s acknowledged", msg.RoutingKey)
}
