package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

// SessionServiceRepository reads live room state from the Study Session
// service (ListActiveSessions() / GetParticipants()), which owns it; admin_db
// only stores the moderation audit log.
type SessionServiceRepository struct {
	baseURL    string
	httpClient *http.Client
}

func NewSessionService(baseURL string) *SessionServiceRepository {
	return &SessionServiceRepository{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type studySessionDTO struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	CreatorID        string     `json:"creator_id"`
	ParticipantLimit int        `json:"participant_limit"`
	ParticipantCount int        `json:"participant_count"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	EndedAt          *time.Time `json:"ended_at"`
}

type participantDTO struct {
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	JoinedAt    time.Time  `json:"joined_at"`
	LeftAt      *time.Time `json:"left_at"`
}

func (r *SessionServiceRepository) ListActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	var raw []studySessionDTO
	if err := r.getJSON(ctx, "/api/v1/study-session/active", &raw); err != nil {
		return nil, err
	}
	sessions := make([]domain.SessionOverview, len(raw))
	for i, s := range raw {
		sessions[i] = domain.SessionOverview{
			SessionID:        s.ID,
			Title:            s.Name,
			HostID:           s.CreatorID,
			IsActive:         s.Status == "ACTIVE",
			MaxParticipants:  s.ParticipantLimit,
			ParticipantCount: s.ParticipantCount,
			CreatedAt:        s.CreatedAt,
			EndedAt:          s.EndedAt,
		}
	}
	return sessions, nil
}

// GetActiveSession looks the room up in the active list: Study Session has no
// single-room REST lookup, and its participants endpoint answers for any ID.
func (r *SessionServiceRepository) GetActiveSession(ctx context.Context, sessionID string) (*domain.SessionOverview, error) {
	sessions, err := r.ListActiveSessions(ctx)
	if err != nil {
		return nil, err
	}
	for i := range sessions {
		if sessions[i].SessionID == sessionID {
			return &sessions[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *SessionServiceRepository) ListParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error) {
	var raw []participantDTO
	if err := r.getJSON(ctx, "/api/v1/study-session/participants?session_id="+url.QueryEscape(sessionID), &raw); err != nil {
		return nil, err
	}
	parts := make([]domain.SessionParticipantOverview, len(raw))
	for i, p := range raw {
		parts[i] = domain.SessionParticipantOverview{
			SessionID: sessionID,
			UserID:    p.UserID,
			JoinedAt:  p.JoinedAt,
			LeftAt:    p.LeftAt,
		}
		if p.DisplayName != "" {
			name := p.DisplayName
			parts[i].DisplayName = &name
		}
	}
	return parts, nil
}

func (r *SessionServiceRepository) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrUpstream, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domain.ErrNotFound
	case resp.StatusCode == http.StatusBadRequest:
		return fmt.Errorf("%w: GET %s rejected by study session", domain.ErrInvalid, path)
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("%w: GET %s returned %d", domain.ErrUpstream, path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("%w: decode %s: %v", domain.ErrUpstream, path, err)
	}
	return nil
}
