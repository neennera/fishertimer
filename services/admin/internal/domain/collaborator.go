package domain

import "context"

// SessionClient sends moderation commands to Study Session.
type SessionClient interface {
	// KickParticipant runs Study Session LeaveSession() for the user.
	// It returns ErrStateChanged when the user had already left.
	KickParticipant(ctx context.Context, sessionID, userID string) (sessionEnded bool, err error)
	// CloseSession runs Study Session EndSession() for the room.
	CloseSession(ctx context.Context, sessionID string) error
}
