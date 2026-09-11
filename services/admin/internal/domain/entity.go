package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("admin: resource not found")
	ErrInvalid  = errors.New("admin: invalid input")
)

type ModerationReport struct {
	ID         string    `json:"id"`
	ReporterID string    `json:"reporter_id"`
	ReportedID string    `json:"reported_id"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
