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

// awardPayload is POST /api/v1/reward/award (UC-05 S-2 / UC-09): everything
// Reward needs to roll the catch. cycle_id makes a retried call harmless.
type awardPayload struct {
	UserID           string `json:"user_id"`
	SessionID        string `json:"session_id"`
	CycleID          string `json:"cycle_id"`
	WorkMinutes      int    `json:"work_duration"`
	ParticipantCount int    `json:"participant_count"`
	Reason           string `json:"reason"`
}

func (c *HTTPRewardClient) AwardReward(ctx context.Context, req domain.AwardRequest) error {
	body, err := json.Marshal(awardPayload{
		UserID:           req.UserID,
		SessionID:        req.SessionID,
		CycleID:          req.CycleID,
		WorkMinutes:      req.WorkMinutes,
		ParticipantCount: req.ParticipantCount,
		Reason:           "CompleteCycle",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal award payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/reward/award", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create award request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute award request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("reward service returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
