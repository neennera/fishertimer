package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("admin: resource not found")
	ErrInvalid  = errors.New("admin: invalid input")
)

type SessionOverview struct {
	SessionID        string    `json:"session_id"`
	Name             string    `json:"name"`
	CreatorID        string    `json:"creator_id"`
	ParticipantCount int       `json:"participant_count"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type SessionParticipantOverview struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	JoinedAt  time.Time `json:"joined_at"`
}
