package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type HTTPRewardClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRewardClient(baseURL string) domain.RewardClient {
	return &HTTPRewardClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type awardPayload struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

func (c *HTTPRewardClient) AwardReward(ctx context.Context, userID, reason string) error {
	payload := awardPayload{
		UserID: userID,
		Reason: reason,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal award payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/reward/award", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create award request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute award request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("reward service returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
