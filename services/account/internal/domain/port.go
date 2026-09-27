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

// UnlockedReward mirrors the reward service's UnlockedReward - a single catch
// of a catalog item, pulled by user_id. A user can catch the same item more
// than once, so this is one entry per catch, not per distinct item.
type UnlockedReward struct {
	UserRewardID string    `json:"user_reward_id"`
	ItemID       string    `json:"item_id"`
	UserID       string    `json:"user_id"`
	CycleID      string    `json:"cycle_id"`
	ItemName     string    `json:"item_name"`
	Category     string    `json:"category"`
	Rarity       string    `json:"rarity"`
	BaseWeight   float64   `json:"base_weight"`
	ScoreValue   int       `json:"score_value"`
	AssetURL     string    `json:"asset_url"`
	AwardedAt    time.Time `json:"awarded_at"`
}

// RewardClient is the driven port for the Reward service collaborator.
// Implemented by internal/adapter/client.
type RewardClient interface {
	GetRewards(ctx context.Context, userID string) ([]UnlockedReward, error)
}

// RewardSummaryItem is one distinct catalog item a user has unlocked, with
// how many times they have caught/earned it.
type RewardSummaryItem struct {
	ItemName   string `json:"name"`
	Rarity     string `json:"rarity"`
	AssetURL   string `json:"asset_url"`
	ItemType   string `json:"type"`
	ScoreValue int    `json:"score_value"`
	Count      int    `json:"count"`
}

// RewardsSummary is the account service's shaped view of a user's rewards:
// totals plus one entry per distinct item, grouped from the Reward service's
// flat per-catch list. TotalScore sums ScoreValue across every catch
// (including repeats) - the catalog has no "cost" field, so this replaces
// what would otherwise be a total-cost figure.
type RewardsSummary struct {
	TotalAwardsEarned int                 `json:"total_awards_earned"`
	TotalScore        int                 `json:"total_score"`
	Items             []RewardSummaryItem `json:"items"`
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
