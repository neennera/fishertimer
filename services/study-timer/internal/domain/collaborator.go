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
	// DisplayName is stored on the reward rows; Leaderboard shows it beside
	// the user's rank. Empty when the user is no longer in the room.
	DisplayName string
}

type RewardClient interface {
	AwardReward(ctx context.Context, req AwardRequest) error
}

// RoomSnapshot is what a completed work cycle needs from the room.
type RoomSnapshot struct {
	ParticipantCount int
	// DisplayName is the user's name in the room roster, or "" when they
	// are no longer in it.
	DisplayName string
}

// SessionClient reads the room from Study Session (UC-09 step 2).
type SessionClient interface {
	Room(ctx context.Context, sessionID, userID string) (RoomSnapshot, error)
}
