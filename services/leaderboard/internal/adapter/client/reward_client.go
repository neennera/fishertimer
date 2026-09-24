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

func (c *HTTPRewardClient) ViewRewards(ctx context.Context, userID string) ([]domain.FishReward, error) {
	reqURL := fmt.Sprintf("%s/api/v1/reward/rewards?user_id=%s", c.baseURL, url.QueryEscape(userID))

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
