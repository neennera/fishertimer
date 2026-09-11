package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("auth: resource not found")
	ErrInvalid  = errors.New("auth: invalid input")
)

type UserAccount struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
