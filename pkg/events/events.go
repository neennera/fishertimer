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
	// ReasonIdleTimeout: no running work cycle for IDLE_TIMEOUT_MINUTES
	// (UC-03 E-7 / UC-05 E-8).
	ReasonIdleTimeout = "IDLE_TIMEOUT"
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
	Reason     string    `json:"reason"` // LEFT | KICKED | DISCONNECT_TIMEOUT | IDLE_TIMEOUT
}

// SessionEnded is published by Study Session when all members leave or admin force-closes the room.
type SessionEnded struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	SessionID  string    `json:"session_id"`
	Reason     string    `json:"reason"` // EMPTY | ADMIN_CLOSED | TIMEOUT_24H
}

// RabbitMQ topology for profile changes: Account publishes, Study Session and
// Leaderboard each consume their own queue, so a rename shows up in active
// rooms and on the leaderboard without the user signing in again (UC-07 S-2).
const (
	AccountExchange = "fisher.account"
	AccountDLX      = "fisher.account.dlx"

	// One queue per consumer, each bound to the profile routing key, so both
	// get every event and a slow one never holds the other back.
	SessionProfileQueue     = "session.profile-events"
	SessionProfileDLQ       = "session.profile-events.dlq"
	LeaderboardProfileQueue = "leaderboard.profile-events"
	LeaderboardProfileDLQ   = "leaderboard.profile-events.dlq"

	RoutingKeyProfileUpdated = "profile.updated"
)

// ProfileUpdated is published by Account after a user changes their display name.
type ProfileUpdated struct {
	EventID     string    `json:"event_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
}
