package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// Renamer is what the consumer needs from the usecase.
type Renamer interface {
	RenameParticipant(ctx context.Context, userID, displayName string) error
}

// ProfileConsumer applies Account's profile.updated events to the room the user
// is in, so a rename shows in the roster without re-joining.
type ProfileConsumer struct {
	url string
	uc  Renamer

	// failures counts transient failures in a row, to pace the requeues.
	failures int
}

func NewProfileConsumer(url string, uc Renamer) *ProfileConsumer {
	return &ProfileConsumer{url: url, uc: uc}
}

// Run consumes session.profile-events until ctx is cancelled, reconnecting with
// backoff whenever the connection drops; meanwhile events wait in the durable queue.
func (c *ProfileConsumer) Run(ctx context.Context) {
	backoff := time.Second
	for {
		err := c.consumeOnce(ctx)
		if ctx.Err() != nil {
			log.Printf("study-session: stopping profile consumer")
			return
		}
		if err == nil {
			backoff = time.Second
		}
		log.Printf("study-session: profile consumer disconnected (%v); reconnecting in %s", err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, 30*time.Second)
	}
}

func (c *ProfileConsumer) consumeOnce(ctx context.Context) error {
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
	if err := DeclareProfileTopology(ch); err != nil {
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(events.SessionProfileQueue, "study-session-profile-consumer", false, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("study-session: profile consumer started on queue %s", events.SessionProfileQueue)

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
			c.handle(ctx, msg)
		}
	}
}

// Verdict is what to do with a delivery.
type Verdict int

const (
	Ack     Verdict = iota // handled, or nothing to do
	Reject                 // can never succeed: dead-letter it
	Requeue                // transient failure: try again later
)

// Handle decides what to do with one profile.updated body. It is separate from
// the AMQP plumbing so it can be tested without a broker.
func Handle(ctx context.Context, uc Renamer, routingKey string, body []byte) Verdict {
	if routingKey != events.RoutingKeyProfileUpdated {
		return Ack // not ours
	}
	var ev events.ProfileUpdated
	if err := json.Unmarshal(body, &ev); err != nil {
		return Reject
	}
	if ev.UserID == "" || ev.DisplayName == "" {
		return Reject
	}
	err := uc.RenameParticipant(ctx, ev.UserID, ev.DisplayName)
	switch {
	case err == nil:
		return Ack
	case errors.Is(err, domain.ErrInvalid):
		return Reject
	default:
		return Requeue
	}
}

func (c *ProfileConsumer) handle(ctx context.Context, msg amqp.Delivery) {
	verdict := Handle(ctx, c.uc, msg.RoutingKey, msg.Body)
	if verdict != Requeue {
		c.failures = 0
	}
	switch verdict {
	case Ack:
		_ = msg.Ack(false)
	case Reject:
		log.Printf("study-session: %s rejected (unusable payload), dead-lettering", msg.RoutingKey)
		_ = msg.Nack(false, false)
	case Requeue:
		c.failures++
		wait := backoffFor(c.failures)
		log.Printf("study-session: transient error on %s, requeueing in %s", msg.RoutingKey, wait)
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
		_ = msg.Nack(false, true)
	}
}

// backoffFor is how long to wait before putting a message back after its
// failures-th failure in a row: 1s, 2s, 4s ... capped at 30s. Without a pause a
// down database would make the broker redeliver the same message at network
// speed, pinning a core and blocking every later rename behind it.
func backoffFor(failures int) time.Duration {
	if failures < 1 {
		return 0
	}
	return min(time.Second<<min(failures-1, 5), 30*time.Second)
}
