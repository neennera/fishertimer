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
}
