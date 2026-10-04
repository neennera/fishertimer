package domain

import "context"

// AwardRequest is one completed work cycle sent to Reward (UC-05 S-2,
// UC-09). Reward deduplicates by CycleID, so it is safe to send again.
type AwardRequest struct {
	UserID           string
	SessionID        string
	CycleID          string
	WorkMinutes      int
	ParticipantCount int
}

type RewardClient interface {
	AwardReward(ctx context.Context, req AwardRequest) error
}

// SessionClient reads the room from Study Session (UC-09 step 2).
type SessionClient interface {
	ParticipantCount(ctx context.Context, sessionID string) (int, error)
}
