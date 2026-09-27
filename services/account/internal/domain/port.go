package domain

import (
	"context"
	"time"
)

// OAuthProvider is the driven port for Google sign-in.
// Implemented by internal/adapter/oauth.
type OAuthProvider interface {
	// AuthCodeURL builds the Google consent-screen URL, tagged with an
	// anti-CSRF state value.
	AuthCodeURL(state string) string

	// FetchProfile exchanges the one-time authorization code for the user's
	// Google profile.
	FetchProfile(ctx context.Context, code string) (*GoogleProfile, error)
}

// TimerStatistics mirrors the study-timer service's TimerHistory - what a
// user has done with the Pomodoro timer, pulled by user_id.
type TimerStatistics struct {
	UserID            string    `json:"user_id"`
	SessionsJoined    int       `json:"sessions_joined"`
	CyclesCompleted   int       `json:"cycles_completed"`
	TotalFocusMinutes int       `json:"total_focus_minutes"`
	LastActive        time.Time `json:"last_active"`
}

// TimerClient is the driven port for the Study Timer service collaborator.
// Implemented by internal/adapter/client.
type TimerClient interface {
	GetStatistics(ctx context.Context, userID string) (*TimerStatistics, error)
}

// TokenService is the driven port for the session JWT.
// Implemented by internal/adapter/token.
type TokenService interface {
	Issue(user *UserAccount) (token string, claims *TokenClaims, err error)
	Verify(token string) (*TokenClaims, error)

	// IssueSignUp / VerifySignUp handle the short-lived ticket that carries a
	// verified Google identity to the sign-up form. A sign-up ticket is never
	// accepted as a session, and a session is never accepted as a ticket.
	IssueSignUp(profile *GoogleProfile) (token string, ticket *SignUpTicket, err error)
	VerifySignUp(token string) (*SignUpTicket, error)
}
