module github.com/neennera/fishertimer/services/account

go 1.25.0

require (
	github.com/lib/pq v1.10.9
	github.com/neennera/fishertimer/pkg/discovery v0.0.0
	github.com/neennera/fishertimer/pkg/events v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
)

replace github.com/neennera/fishertimer/pkg/events => ../../pkg/events

replace github.com/neennera/fishertimer/pkg/discovery => ../../pkg/discovery
