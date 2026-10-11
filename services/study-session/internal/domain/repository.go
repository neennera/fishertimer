package domain

import (
	"context"
	"time"
)

// Repository persists rooms, participants and the event outbox. Each
// command method is one transaction: the state change and the events it
// produces commit together or not at all.
type Repository interface {
	// CreateSession inserts the room, enrolls the creator as its first
	// participant and queues the joined event. ErrAlreadyInSession when the
	// creator is already in a room.
	CreateSession(ctx context.Context, s *StudySession, creator *Participant, joined OutboxEvent) error

	// JoinSession adds p to its room, checking capacity atomically.
	// ErrNotFound, ErrSessionEnded, ErrSessionFull or ErrAlreadyInSession.
	// Returns the room as it is after the join.
	JoinSession(ctx context.Context, p *Participant, joined OutboxEvent) (*StudySession, error)

	// LeaveSession closes the user's participation. When it empties the room
	// the room is ended with reason EMPTY and ended is queued as well; left
	// is queued whenever the user actually left. A user who is not an active
	// participant is a no-op (Left = false), not an error.
	LeaveSession(ctx context.Context, sessionID, userID, reason string, now time.Time, left, ended OutboxEvent) (*LeaveResult, error)

	// EndSession ends an ACTIVE room, closes every open participation and
	// queues ended. Returns the room and whether it was ended by this call
	// (false when it had already ended). ErrNotFound for an unknown room.
	EndSession(ctx context.Context, sessionID, reason string, now time.Time, ended OutboxEvent) (*StudySession, bool, error)

	GetSession(ctx context.Context, id string) (*StudySession, error)
	ListActiveSessions(ctx context.Context) ([]StudySession, error)
	// GetParticipants returns the room's active participants, oldest first.
	GetParticipants(ctx context.Context, sessionID string) ([]Participant, error)
	// GetActiveParticipation returns the user's open participation, or
	// ErrNotFound when they are not in any room.
	GetActiveParticipation(ctx context.Context, userID string) (*Participant, error)
	// GetLastParticipation returns the user's most recent participation in
	// the room (open or closed), or ErrNotFound.
	GetLastParticipation(ctx context.Context, sessionID, userID string) (*Participant, error)

	// TouchParticipant records a heartbeat. Reports false when the user is
	// not an active participant of the room.
	TouchParticipant(ctx context.Context, sessionID, userID string, now time.Time) (bool, error)
	// ListStaleParticipants returns active participants whose last heartbeat
	// is before cutoff.
	ListStaleParticipants(ctx context.Context, cutoff time.Time) ([]Participant, error)
	// ListSessionsCreatedBefore returns ACTIVE rooms created before cutoff.
	ListSessionsCreatedBefore(ctx context.Context, cutoff time.Time) ([]StudySession, error)

	// PendingEvents returns up to limit unpublished outbox events, oldest
	// first.
	PendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkEventPublished(ctx context.Context, id int64, now time.Time) error
	MarkEventFailed(ctx context.Context, id int64, cause string) error
}

// EventPublisher sends one event to the message broker and returns only
// once the broker has confirmed it.
type EventPublisher interface {
	Publish(ctx context.Context, routingKey, eventID string, payload []byte) error
}

// TimerReader reads participants' timers from Study Timer.
type TimerReader interface {
	RoomTimers(ctx context.Context, sessionID string) ([]TimerActivity, error)
}
