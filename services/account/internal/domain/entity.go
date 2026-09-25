package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("account: resource not found")
	ErrInvalid      = errors.New("account: invalid input")
	ErrUnauthorized = errors.New("account: unauthorized")
)

// Roles (UC-06 E-4): Google sign-in can only ever produce CUSTOMER.
// ADMIN rows are pre-provisioned directly in the database.
const (
	RoleCustomer = "CUSTOMER"
	RoleAdmin    = "ADMIN"
)

// UserAccount mirrors account_db.users (docs/database/schema.dbml).
// Accounts are identified by e-mail, as UC-06 specifies.
type UserAccount struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GoogleProfile is what we read back from Google after a successful sign-in.
type GoogleProfile struct {
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// Valid reports whether the profile may be used to sign in. An unverified
// e-mail must be refused: accounts are matched by e-mail, so accepting one
// would let somebody claim another person's account.
func (p *GoogleProfile) Valid() bool {
	return p != nil && strings.TrimSpace(p.Email) != "" && p.EmailVerified
}

// TokenClaims is the decoded session JWT: user_id and role, as UC-06 requires.
type TokenClaims struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Session is the result of a completed sign-in.
type Session struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *UserAccount `json:"user"`
}

// SignUpTicket is a Google identity that has no account yet. It travels from
// the OAuth callback to the display-name form inside a signed, short-lived
// token, so the browser can read it back but cannot change the e-mail.
type SignUpTicket struct {
	Email     string    `json:"email"`
	Picture   string    `json:"picture"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SignInResult is what the OAuth callback produces: a Session when the e-mail
// already has an account, otherwise a sign-up ticket to finish with a display
// name. Exactly one of the two is set.
type SignInResult struct {
	Session     *Session
	SignUpToken string
	SignUp      *SignUpTicket
}

// MaxDisplayNameLength matches users.display_name VARCHAR(100).
const MaxDisplayNameLength = 100
