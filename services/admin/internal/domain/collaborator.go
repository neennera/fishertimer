package domain

import "context"

type SessionClient interface {
	KickParticipant(ctx context.Context, sessionID, userID string) error
	CloseSession(ctx context.Context, sessionID string) error
}
