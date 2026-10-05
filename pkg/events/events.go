package events

import "time"

// RabbitMQ Topology constants for Study Session <-> Study Timer messaging.
const (
	SessionExchange   = "fisher.session"
	SessionDLX        = "fisher.session.dlx"
	TimerSessionQueue = "timer.session-events"
	TimerSessionDLQ   = "timer.session-events.dlq"

	RoutingKeyJoined  = "session.participant.joined"
	RoutingKeyLeft    = "session.participant.left"
	RoutingKeyEnded   = "session.ended"
	BindingSessionAll = "session.#"
)

// Reason constants for ParticipantLeft
const (
	ReasonLeft              = "LEFT"
	ReasonKicked            = "KICKED"
	ReasonDisconnectTimeout = "DISCONNECT_TIMEOUT"
)

// Reason constants for SessionEnded
const (
	ReasonEmpty       = "EMPTY"
	ReasonAdminClosed = "ADMIN_CLOSED"
	ReasonTimeout24H  = "TIMEOUT_24H"
)

// ParticipantJoined is published by Study Session when a user creates or joins a room.
type ParticipantJoined struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
}

// ParticipantLeft is published by Study Session when a user leaves, is kicked, or times out.
type ParticipantLeft struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
	Reason     string    `json:"reason"` // LEFT | KICKED | DISCONNECT_TIMEOUT
}

// SessionEnded is published by Study Session when all members leave or admin force-closes the room.
type SessionEnded struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	SessionID  string    `json:"session_id"`
	Reason     string    `json:"reason"` // EMPTY | ADMIN_CLOSED | TIMEOUT_24H
}
