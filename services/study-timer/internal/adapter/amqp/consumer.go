package amqp

import (
	"context"
	"encoding/json"
	"log"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch *amqp.Channel
	uc usecase.Usecase
}

func NewConsumer(ch *amqp.Channel, uc usecase.Usecase) *Consumer {
	return &Consumer{
		ch: ch,
		uc: uc,
	}
}

// Start begins consuming events from timer.session-events queue.
func (c *Consumer) Start(ctx context.Context) error {
	// Set prefetch to 10 for fair dispatch
	if err := c.ch.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := c.ch.Consume(
		events.TimerSessionQueue,
		"study-timer-consumer",
		false, // manual ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("study-timer: RabbitMQ consumer worker started on queue %s", events.TimerSessionQueue)

	for {
		select {
		case <-ctx.Done():
			log.Printf("study-timer: stopping RabbitMQ consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("study-timer: RabbitMQ delivery channel closed")
				return nil
			}
			c.handleDelivery(ctx, msg)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, msg amqp.Delivery) {
	switch msg.RoutingKey {
	case events.RoutingKeyLeft:
		var ev events.ParticipantLeft
		if err := json.Unmarshal(msg.Body, &ev); err != nil {
			log.Printf("[Consumer] JSON parse error for %s: %v. Nacking to DLQ", msg.RoutingKey, err)
			_ = msg.Nack(false, false) // Requeue = false -> routes to DLQ
			return
		}
		if ev.EventID == "" || ev.SessionID == "" || ev.UserID == "" {
			log.Printf("[Consumer] Malformed event payload %s. Nacking to DLQ", msg.RoutingKey)
			_ = msg.Nack(false, false)
			return
		}

		log.Printf("[Consumer] Processing participant left: event_id=%s, session_id=%s, user_id=%s, reason=%s",
			ev.EventID, ev.SessionID, ev.UserID, ev.Reason)

		if err := c.uc.FinalizeParticipantTimer(ctx, ev.SessionID, ev.UserID, ev.EventID, ev.Reason); err != nil {
			log.Printf("[Consumer] Transient DB error finalizing timer: %v. Requeueing", err)
			_ = msg.Nack(false, true) // Requeue on transient DB error
			return
		}

		_ = msg.Ack(false)
		log.Printf("[Consumer] Participant left event %s acknowledged", ev.EventID)

	case events.RoutingKeyEnded:
		var ev events.SessionEnded
		if err := json.Unmarshal(msg.Body, &ev); err != nil {
			log.Printf("[Consumer] JSON parse error for %s: %v. Nacking to DLQ", msg.RoutingKey, err)
			_ = msg.Nack(false, false)
			return
		}
		if ev.EventID == "" || ev.SessionID == "" {
			log.Printf("[Consumer] Malformed event payload %s. Nacking to DLQ", msg.RoutingKey)
			_ = msg.Nack(false, false)
			return
		}

		log.Printf("[Consumer] Processing session ended: event_id=%s, session_id=%s, reason=%s",
			ev.EventID, ev.SessionID, ev.Reason)

		if err := c.uc.FinalizeSessionTimers(ctx, ev.SessionID, ev.EventID, ev.Reason); err != nil {
			log.Printf("[Consumer] Transient DB error finalizing session timers: %v. Requeueing", err)
			_ = msg.Nack(false, true)
			return
		}

		_ = msg.Ack(false)
		log.Printf("[Consumer] Session ended event %s acknowledged", ev.EventID)

	case events.RoutingKeyJoined:
		var ev events.ParticipantJoined
		if err := json.Unmarshal(msg.Body, &ev); err != nil {
			log.Printf("[Consumer] JSON parse error for %s: %v. Nacking to DLQ", msg.RoutingKey, err)
			_ = msg.Nack(false, false)
			return
		}
		log.Printf("[Consumer] Participant joined: session=%s, user=%s (W2 feature - acknowledged)",
			ev.SessionID, ev.UserID)
		_ = msg.Ack(false)

	default:
		log.Printf("[Consumer] Unknown routing key %s. Acknowledging and skipping", msg.RoutingKey)
		_ = msg.Ack(false)
	}
}
