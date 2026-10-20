package amqp

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/neennera/fishertimer/pkg/events"
)

// consumerQueue is one consumer's queue with its dead-letter queue.
type consumerQueue struct{ name, dlq string }

var consumers = []consumerQueue{
	{events.SessionProfileQueue, events.SessionProfileDLQ},
	{events.LeaderboardProfileQueue, events.LeaderboardProfileDLQ},
}

// DeclareTopology declares the fisher.account exchange together with each
// consumer's queue, dead-letter exchange and DLQ. Declaring is idempotent.
//
// The publisher declares the consumers' queues too, because RabbitMQ drops a
// message published to an exchange with no bound queue: a rename made before
// Study Session or Leaderboard ever started would be lost. With the queues
// declared here, those events wait for them.
//
// The arguments must stay identical to the consumers' own declarations
// (services/study-session and services/leaderboard adapter/amqp/topology.go);
// RabbitMQ rejects a re-declare with different arguments (PRECONDITION_FAILED).
func DeclareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(events.AccountExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", events.AccountExchange, err)
	}
	if err := ch.ExchangeDeclare(events.AccountDLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", events.AccountDLX, err)
	}
	for _, q := range consumers {
		if _, err := ch.QueueDeclare(q.dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue %s: %w", q.dlq, err)
		}
		if err := ch.QueueBind(q.dlq, q.dlq, events.AccountDLX, false, nil); err != nil {
			return fmt.Errorf("bind queue %s: %w", q.dlq, err)
		}
		if _, err := ch.QueueDeclare(q.name, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange":    events.AccountDLX,
			"x-dead-letter-routing-key": q.dlq,
		}); err != nil {
			return fmt.Errorf("declare queue %s: %w", q.name, err)
		}
		if err := ch.QueueBind(q.name, events.RoutingKeyProfileUpdated, events.AccountExchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s: %w", q.name, err)
		}
	}
	return nil
}
