// Package amqp receives Account's profile.updated events for the leaderboard.
package amqp

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/neennera/fishertimer/pkg/events"
)

// DeclareProfileTopology declares the fisher.account exchange and this
// service's own queue for profile changes, with its dead-letter exchange and
// DLQ. Declaring is idempotent.
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
	if _, err := ch.QueueDeclare(events.LeaderboardProfileDLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", events.LeaderboardProfileDLQ, err)
	}
	if err := ch.QueueBind(events.LeaderboardProfileDLQ, events.LeaderboardProfileDLQ, events.AccountDLX, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", events.LeaderboardProfileDLQ, err)
	}
	if _, err := ch.QueueDeclare(events.LeaderboardProfileQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    events.AccountDLX,
		"x-dead-letter-routing-key": events.LeaderboardProfileDLQ,
	}); err != nil {
		return fmt.Errorf("declare queue %s: %w", events.LeaderboardProfileQueue, err)
	}
	if err := ch.QueueBind(events.LeaderboardProfileQueue, events.RoutingKeyProfileUpdated, events.AccountExchange, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", events.LeaderboardProfileQueue, err)
	}
	return nil
}
