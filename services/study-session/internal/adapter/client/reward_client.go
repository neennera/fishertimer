package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/study-session/internal/domain"
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

func (c *HTTPRewardClient) AwardReward(ctx context.Context, userID, reason string) error {
	payload, err := json.Marshal(map[string]string{
		"user_id": userID,
		"reason":  reason,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/reward/award", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("reward service returned status %d", resp.StatusCode)
	}

	return nil
}
