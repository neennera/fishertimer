package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

type HTTPSessionClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSessionClient(baseURL string) domain.SessionClient {
	return &HTTPSessionClient{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *HTTPSessionClient) KickParticipant(ctx context.Context, sessionID, userID string) error {
	payload, _ := json.Marshal(map[string]string{"session_id": sessionID, "user_id": userID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/study-session/leave", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("session service returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPSessionClient) CloseSession(ctx context.Context, sessionID string) error {
	payload, _ := json.Marshal(map[string]string{"session_id": sessionID, "user_id": "admin"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/study-session/end", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("session service returned status %d", resp.StatusCode)
	}
	return nil
}
