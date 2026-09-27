// Package client holds driven adapters for the account service's downstream
// collaborators. HTTPTimerClient calls the Study Timer service directly
// (no cross-module import - each service is its own Go module per go.work),
// mirroring services/study-timer/internal/adapter/client.HTTPRewardClient.
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

type HTTPTimerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewTimerClient(baseURL string) domain.TimerClient {
	return &HTTPTimerClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *HTTPTimerClient) GetStatistics(ctx context.Context, userID string) (*domain.TimerStatistics, error) {
	target := c.baseURL + "/api/v1/study-timer/statistics?" + url.Values{"user_id": {userID}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("timer client: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("timer client: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("timer client: study-timer returned status %d", resp.StatusCode)
	}

	var stats domain.TimerStatistics
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("timer client: decode response: %w", err)
	}
	return &stats, nil
}
