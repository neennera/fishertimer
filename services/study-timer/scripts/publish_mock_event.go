package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/neennera/fishertimer/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	eventType := flag.String("type", "left", "Event type: left | end | join")
	sessionID := flag.String("session", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "Session UUID")
	userID := flag.String("user", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "User UUID")
	eventID := flag.String("event", "", "Event ID (defaults to new random ID)")
	rabbitURL := flag.String("url", "amqp://admin:adminpassword@localhost:5672/", "RabbitMQ URL")
	corrupt := flag.Bool("corrupt", false, "Publish corrupt/invalid payload to test DLQ")
	flag.Parse()

	if *eventID == "" {
		*eventID = fmt.Sprintf("evt-%x", time.Now().UnixNano())
	}

	conn, err := amqp.Dial(*rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ at %s: %v", *rabbitURL, err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	var routingKey string
	var body []byte

	if *corrupt {
		routingKey = events.RoutingKeyLeft
		body = []byte("THIS_IS_INVALID_CORRUPT_JSON_DATA_FOR_DLQ_TEST")
		fmt.Printf("--> Sending POISON/CORRUPT message to exchange: %s with routingKey: %s\n",
			events.SessionExchange, routingKey)
	} else {
		switch *eventType {
		case "left":
			routingKey = events.RoutingKeyLeft
			ev := events.ParticipantLeft{
				EventID:    *eventID,
				OccurredAt: time.Now().UTC(),
				SessionID:  *sessionID,
				UserID:     *userID,
				Reason:     events.ReasonLeft,
			}
			body, _ = json.Marshal(ev)
		case "end":
			routingKey = events.RoutingKeyEnded
			ev := events.SessionEnded{
				EventID:    *eventID,
				OccurredAt: time.Now().UTC(),
				SessionID:  *sessionID,
				Reason:     events.ReasonEmpty,
			}
			body, _ = json.Marshal(ev)
		case "join":
			routingKey = events.RoutingKeyJoined
			ev := events.ParticipantJoined{
				EventID:    *eventID,
				OccurredAt: time.Now().UTC(),
				SessionID:  *sessionID,
				UserID:     *userID,
			}
			body, _ = json.Marshal(ev)
		default:
			log.Fatalf("Unknown event type: %s", *eventType)
		}
		fmt.Printf("--> Publishing %s event: %s\n", routingKey, string(body))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		events.SessionExchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	fmt.Println("✓ Message published successfully to RabbitMQ!")
	os.Exit(0)
}
