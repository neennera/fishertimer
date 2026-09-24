package domain

import "context"

type RewardClient interface {
	AwardReward(ctx context.Context, userID, reason string) error
}
