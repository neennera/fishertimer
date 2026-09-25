package domain

import "context"

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
