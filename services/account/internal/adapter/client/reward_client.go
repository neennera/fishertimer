package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// HTTPRewardClient calls the Reward service directly (no cross-module import
// - each service is its own Go module per go.work).
type HTTPRewardClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRewardClient(baseURL string) domain.RewardClient {
	return &HTTPRewardClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *HTTPRewardClient) GetRewards(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	target := c.baseURL + "/api/v1/reward/rewards?" + url.Values{"user_id": {userID}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("reward client: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reward client: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("reward client: reward service returned status %d", resp.StatusCode)
	}

	var rewards []domain.UnlockedReward
	if err := json.NewDecoder(resp.Body).Decode(&rewards); err != nil {
		return nil, fmt.Errorf("reward client: decode response: %w", err)
	}
	return rewards, nil
}
