package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		return false, fmt.Errorf("%w: This participant has already left the room.", domain.ErrStateChanged)
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
		// Study Session's own wording ("study-session: invalid input: ...") is
		// for logs, not for the admin: log it, and send a plain sentence.
		log.Printf("admin: study session answered %d to POST %s: %s", resp.StatusCode, path, upstreamMessage(resp.Body))
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: This room is no longer active.", domain.ErrNotFound)
		case http.StatusConflict:
			return fmt.Errorf("%w: The room changed before the action could finish.", domain.ErrStateChanged)
		case http.StatusBadRequest:
			return fmt.Errorf("%w: The study session service rejected the request.", domain.ErrInvalid)
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
