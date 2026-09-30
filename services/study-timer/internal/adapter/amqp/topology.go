package amqp

import (
	"fmt"
	"log"

	"github.com/neennera/fishertimer/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareTopology sets up the exchanges, queues, DLX, DLQ, and bindings idempotently.
func DeclareTopology(ch *amqp.Channel) error {
	// 1. Declare Main Exchange (topic, durable)
	if err := ch.ExchangeDeclare(
		events.SessionExchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("declare main exchange (%s): %w", events.SessionExchange, err)
	}

	// 2. Declare Dead Letter Exchange (direct, durable)
	if err := ch.ExchangeDeclare(
		events.SessionDLX,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("declare dlx exchange (%s): %w", events.SessionDLX, err)
	}

	// 3. Declare DLQ
	if _, err := ch.QueueDeclare(
		events.TimerSessionDLQ,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("declare dlq (%s): %w", events.TimerSessionDLQ, err)
	}

	// Bind DLQ to DLX
	if err := ch.QueueBind(
		events.TimerSessionDLQ,
		events.TimerSessionDLQ, // routing key matching dlq name
		events.SessionDLX,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind dlq (%s): %w", events.TimerSessionDLQ, err)
	}

	// 4. Declare Main Queue with DLX arguments
	queueArgs := amqp.Table{
		"x-dead-letter-exchange":    events.SessionDLX,
		"x-dead-letter-routing-key": events.TimerSessionDLQ,
	}
	if _, err := ch.QueueDeclare(
		events.TimerSessionQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		queueArgs,
	); err != nil {
		return fmt.Errorf("declare main queue (%s): %w", events.TimerSessionQueue, err)
	}

	// 5. Bind Main Queue to Main Exchange with wildcard routing key
	if err := ch.QueueBind(
		events.TimerSessionQueue,
		events.BindingSessionAll, // "session.#"
		events.SessionExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind main queue to exchange (%s): %w", events.SessionExchange, err)
	}

	log.Printf("study-timer: RabbitMQ topology declared (exchange: %s, queue: %s, dlx: %s, dlq: %s)",
		events.SessionExchange, events.TimerSessionQueue, events.SessionDLX, events.TimerSessionDLQ)
	return nil
}
