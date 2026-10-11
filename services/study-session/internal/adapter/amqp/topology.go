package amqp

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/neennera/fishertimer/pkg/events"
)

// DeclareTopology declares the fisher.session exchange together with Study
// Timer's queue, dead-letter exchange and DLQ. Declaring is idempotent.
//
// The publisher declares the consumer's queue too because RabbitMQ drops a
// message published to an exchange with no bound queue: if Study Session
// published before Study Timer had ever started, those events would be
// lost. With the queue declared here, they wait in timer.session-events
// until the Timer comes up.
//
// The arguments must stay identical to
// services/study-timer/internal/adapter/amqp/topology.go; RabbitMQ rejects
// a re-declare with different arguments (PRECONDITION_FAILED).
func DeclareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(events.SessionExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", events.SessionExchange, err)
	}
	if err := ch.ExchangeDeclare(events.SessionDLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", events.SessionDLX, err)
	}
	if _, err := ch.QueueDeclare(events.TimerSessionDLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", events.TimerSessionDLQ, err)
	}
	if err := ch.QueueBind(events.TimerSessionDLQ, events.TimerSessionDLQ, events.SessionDLX, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", events.TimerSessionDLQ, err)
	}
	if _, err := ch.QueueDeclare(events.TimerSessionQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    events.SessionDLX,
		"x-dead-letter-routing-key": events.TimerSessionDLQ,
	}); err != nil {
		return fmt.Errorf("declare queue %s: %w", events.TimerSessionQueue, err)
	}
	if err := ch.QueueBind(events.TimerSessionQueue, events.BindingSessionAll, events.SessionExchange, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", events.TimerSessionQueue, err)
	}
	return nil
}
