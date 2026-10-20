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

// DeclareProfileTopology declares the fisher.account exchange and this service's
// own queue for profile changes, with its dead-letter exchange and DLQ.
// Declaring is idempotent.
//
// The arguments must stay identical to the publisher's declaration in
// services/account/internal/adapter/amqp/topology.go; RabbitMQ rejects a
// re-declare with different arguments (PRECONDITION_FAILED).
func DeclareProfileTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(events.AccountExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", events.AccountExchange, err)
	}
	if err := ch.ExchangeDeclare(events.AccountDLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", events.AccountDLX, err)
	}
	if _, err := ch.QueueDeclare(events.SessionProfileDLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", events.SessionProfileDLQ, err)
	}
	if err := ch.QueueBind(events.SessionProfileDLQ, events.SessionProfileDLQ, events.AccountDLX, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", events.SessionProfileDLQ, err)
	}
	if _, err := ch.QueueDeclare(events.SessionProfileQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    events.AccountDLX,
		"x-dead-letter-routing-key": events.SessionProfileDLQ,
	}); err != nil {
		return fmt.Errorf("declare queue %s: %w", events.SessionProfileQueue, err)
	}
	if err := ch.QueueBind(events.SessionProfileQueue, events.RoutingKeyProfileUpdated, events.AccountExchange, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", events.SessionProfileQueue, err)
	}
	return nil
}
