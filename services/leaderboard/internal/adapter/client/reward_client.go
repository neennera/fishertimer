package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

// HTTPRewardClient implements domain.RewardClient by calling the Reward microservice.
type HTTPRewardClient struct {
	baseURL string
	client  *http.Client
}

func NewRewardClient(baseURL string) domain.RewardClient {
	return &HTTPRewardClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// ViewRewards fetches rewards for a single user.
func (c *HTTPRewardClient) ViewRewards(ctx context.Context, userID string) ([]domain.FishReward, error) {
	reqURL := fmt.Sprintf("%s/api/v1/reward/rewards?user_id=%s", c.baseURL, url.QueryEscape(userID))
	return c.fetchRewards(ctx, reqURL)
}

// ViewAllRewards fetches all rewards across all users for leaderboard ranking (S-3).
func (c *HTTPRewardClient) ViewAllRewards(ctx context.Context) ([]domain.FishReward, error) {
	reqURL := fmt.Sprintf("%s/api/v1/reward/all-rewards", c.baseURL)
	return c.fetchRewards(ctx, reqURL)
}

// GetLastUpdate fetches reward_last_update timestamp for cache validation (S-1).
func (c *HTTPRewardClient) GetLastUpdate(ctx context.Context) (time.Time, error) {
	reqURL := fmt.Sprintf("%s/api/v1/reward/last-update", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return time.Time{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return time.Time{}, fmt.Errorf("reward service returned status %d", resp.StatusCode)
	}

	var body struct {
		RewardLastUpdate string `json:"reward_last_update"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return time.Time{}, err
	}

	ts, err := time.Parse(time.RFC3339Nano, body.RewardLastUpdate)
	if err != nil {
		ts, err = time.Parse(time.RFC3339, body.RewardLastUpdate)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse reward_last_update: %w", err)
		}
	}

	return ts, nil
}

// fetchRewards is a shared helper for both ViewRewards and ViewAllRewards.
func (c *HTTPRewardClient) fetchRewards(ctx context.Context, reqURL string) ([]domain.FishReward, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("reward service returned status %d", resp.StatusCode)
	}

	var rewards []domain.FishReward
	if err := json.NewDecoder(resp.Body).Decode(&rewards); err != nil {
		return nil, err
	}

	return rewards, nil
}
