package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("study-session: resource not found")
	ErrInvalid  = errors.New("study-session: invalid input")
)

type StudySession struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	CreatorID        string    `json:"creator_id"`
	ParticipantLimit int       `json:"participant_limit"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}
