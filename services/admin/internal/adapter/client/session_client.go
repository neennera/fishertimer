package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

// Leave / end reasons understood by Study Session (pkg/events). Admin is its
// own Go module and does not import pkg/events, so they are repeated here.
const (
	reasonKicked      = "KICKED"
	reasonAdminClosed = "ADMIN_CLOSED"
)

// HTTPSessionClient sends moderation commands to Study Session's internal
// REST routes. Study Session finalizes the affected timers itself by
// publishing participant.left / session.ended to Study Timer.
type HTTPSessionClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSessionClient(baseURL string) domain.SessionClient {
	return &HTTPSessionClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type leaveResponse struct {
	Status       string `json:"status"`
	SessionEnded bool   `json:"session_ended"`
}

func (c *HTTPSessionClient) KickParticipant(ctx context.Context, sessionID, userID string) (bool, error) {
	var res leaveResponse
	err := c.post(ctx, "/api/v1/study-session/leave", map[string]string{
		"session_id": sessionID,
		"user_id":    userID,
		"reason":     reasonKicked,
	}, &res)
	if err != nil {
		return false, err
	}
	if res.Status == "already_left" {
		return false, fmt.Errorf("%w: participant has already left", domain.ErrStateChanged)
	}
	return res.SessionEnded, nil
}

func (c *HTTPSessionClient) CloseSession(ctx context.Context, sessionID string) error {
	return c.post(ctx, "/api/v1/study-session/end", map[string]string{
		"session_id": sessionID,
		"reason":     reasonAdminClosed,
	}, nil)
}

func (c *HTTPSessionClient) post(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := upstreamMessage(resp.Body)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", domain.ErrNotFound, msg)
		case http.StatusConflict:
			return fmt.Errorf("%w: %s", domain.ErrStateChanged, msg)
		case http.StatusBadRequest:
			return fmt.Errorf("%w: %s", domain.ErrInvalid, msg)
		}
		return fmt.Errorf("%w: POST %s returned %d", domain.ErrUpstream, path, resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("%w: decode %s: %v", domain.ErrUpstream, path, err)
	}
	return nil
}

// upstreamMessage extracts Study Session's {"error": "..."} body.
func upstreamMessage(body io.Reader) string {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(body, 4<<10)).Decode(&e); err != nil || e.Error == "" {
		return "rejected by study session"
	}
	return e.Error
}
