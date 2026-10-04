package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("study-session: resource not found")
	ErrInvalid  = errors.New("study-session: invalid input")
	// ErrSessionEnded is returned when joining a room that is no longer
	// ACTIVE (UC-02 E-1).
	ErrSessionEnded = errors.New("study-session: room has ended")
	// ErrSessionFull is returned when a room is at its participant limit,
	// including when two users race for the last slot (UC-02 E-2).
	ErrSessionFull = errors.New("study-session: room is full")
	// ErrAlreadyInSession is returned when the user is already an active
	// participant of a room (UC-01 E-2, UC-02 E-3).
	ErrAlreadyInSession = errors.New("study-session: user is already in a room")
)

// Room rules (UC-01).
const (
	MinParticipantLimit = 1
	MaxParticipantLimit = 5
	MaxNameLength       = 60
)

type SessionStatus string

const (
	StatusActive SessionStatus = "ACTIVE"
	StatusEnded  SessionStatus = "ENDED"
)

type StudySession struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	CreatorID        string        `json:"creator_id"`
	ParticipantLimit int           `json:"participant_limit"`
	ParticipantCount int           `json:"participant_count"`
	Status           SessionStatus `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	EndedAt          *time.Time    `json:"ended_at,omitempty"`
	EndReason        string        `json:"end_reason,omitempty"`
}

// Participant is one stay of a user in a room. A user who leaves and joins
// again gets a new row; at most one row per user has LeftAt == nil.
type Participant struct {
	SessionID   string     `json:"session_id"`
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	JoinedAt    time.Time  `json:"joined_at"`
	LastSeenAt  time.Time  `json:"last_seen_at"`
	LeftAt      *time.Time `json:"left_at,omitempty"`
	LeaveReason string     `json:"leave_reason,omitempty"`
}

// LeaveResult describes what a LeaveSession call did.
type LeaveResult struct {
	// Left is false when the user was not an active participant, so nothing
	// changed (UC-04 E-4).
	Left bool
	// SessionEnded is true when this leave emptied the room (UC-03 S-2).
	SessionEnded bool
	// Participant is the closed participation, or nil when Left is false.
	Participant *Participant
}

// OutboxEvent is a domain event waiting to be published to the message
// broker. It is written in the same transaction as the state change it
// describes (transactional outbox), so an event is never lost when the
// broker is down and never sent for a change that rolled back.
type OutboxEvent struct {
	ID          int64
	EventID     string
	RoutingKey  string
	Payload     []byte
	CreatedAt   time.Time
	Attempts    int
	PublishedAt *time.Time
}

// TimerActivity is the part of a participant's timer that idle detection
// needs (UC-05 E-8), read from Study Timer.
type TimerActivity struct {
	UserID      string
	Status      string // STOPPED | RUNNING | PAUSED
	Phase       string // WORK | REST
	LastUpdated time.Time
	// DurationSeconds is the length of the current phase.
	DurationSeconds int
	// RemainingSeconds is the time left in the phase when it was read.
	RemainingSeconds int
}
