package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

// DefaultLogLimit caps how many audit logs ViewLogs returns.
const DefaultLogLimit = 100

type Usecase interface {
	ViewActiveSessions(ctx context.Context) ([]domain.SessionOverview, error)
	ViewParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error)
	ViewSessionDetails(ctx context.Context, sessionID string) (*domain.SessionDetail, error)
	KickUser(ctx context.Context, adminID, sessionID, userID, reason string) (*domain.ModerationResult, error)
	EndSession(ctx context.Context, adminID, sessionID, reason string) (*domain.ModerationResult, error)
	ViewLogs(ctx context.Context, action string) ([]domain.AdminLog, error)
}

type service struct {
	sessions      domain.SessionRepository
	logs          domain.AuditLogRepository
	sessionClient domain.SessionClient
}

func New(sessions domain.SessionRepository, logs domain.AuditLogRepository, sessionClient domain.SessionClient) Usecase {
	return &service{
		sessions:      sessions,
		logs:          logs,
		sessionClient: sessionClient,
	}
}

func (s *service) ViewActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	sessions, err := s.sessions.ListActiveSessions(ctx)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(sessions, func(a, b int) bool {
		return sessions[a].CreatedAt.After(sessions[b].CreatedAt)
	})
	return sessions, nil
}

func (s *service) ViewParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error) {
	detail, err := s.ViewSessionDetails(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return detail.Participants, nil
}

// ViewSessionDetails returns a live room with its members, host first.
// A room that is not live is ErrNotFound (UC-04 E-3).
func (s *service) ViewSessionDetails(ctx context.Context, sessionID string) (*domain.SessionDetail, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("%w: A room id is required.", domain.ErrInvalid)
	}
	sess, err := s.sessions.GetActiveSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	parts, err := s.sessions.ListParticipants(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	for i := range parts {
		parts[i].SessionID = sessionID
		parts[i].IsHost = strings.EqualFold(parts[i].UserID, sess.HostID)
		if parts[i].IsHost {
			sess.HostName = parts[i].DisplayName
		}
	}
	sort.SliceStable(parts, func(a, b int) bool {
		return parts[a].IsHost && !parts[b].IsHost
	})
	return &domain.SessionDetail{SessionOverview: *sess, Participants: parts}, nil
}

// KickUser removes a participant through Study Session LeaveSession() and
// records KICK_USER (UC-04 S-2, UC-10 S-1).
func (s *service) KickUser(ctx context.Context, adminID, sessionID, userID, reason string) (*domain.ModerationResult, error) {
	sessionID, userID = strings.TrimSpace(sessionID), strings.TrimSpace(userID)
	if !isUUID(sessionID) || !isUUID(userID) {
		return nil, fmt.Errorf("%w: That room or user id is not valid.", domain.ErrInvalid)
	}
	entry, err := newLog(adminID, domain.ActionKickUser, userID, reason)
	if err != nil {
		return nil, err
	}
	room, err := s.requireLive(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	// A room without its host is left with nobody in charge (host_id would
	// point at someone who is gone), so the host is not removable one by one:
	// close the room instead.
	// EqualFold: isUUID accepts upper case and Postgres compares uuids without case.
	if strings.EqualFold(room.HostID, userID) {
		return nil, fmt.Errorf("%w: The host can't be removed. Close the room instead.", domain.ErrInvalid)
	}

	var ended bool
	err = s.logs.Record(ctx, entry, func(ctx context.Context) error {
		var kickErr error
		ended, kickErr = s.sessionClient.KickParticipant(ctx, sessionID, userID)
		return kickErr
	})
	if err != nil {
		return nil, err
	}
	return &domain.ModerationResult{SessionEnded: ended, Log: *entry}, nil
}

// EndSession force-closes a room through Study Session EndSession() and
// records FORCE_CLOSE_SESSION (UC-04 S-3, UC-10 S-2).
func (s *service) EndSession(ctx context.Context, adminID, sessionID, reason string) (*domain.ModerationResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if !isUUID(sessionID) {
		return nil, fmt.Errorf("%w: That room id is not valid.", domain.ErrInvalid)
	}
	entry, err := newLog(adminID, domain.ActionForceCloseSession, sessionID, reason)
	if err != nil {
		return nil, err
	}
	// Study Session answers 200 for a room that has already ended, so the
	// "already ended" case is detected here instead (UC-10 E-3).
	if _, err := s.requireLive(ctx, sessionID); err != nil {
		return nil, err
	}

	err = s.logs.Record(ctx, entry, func(ctx context.Context) error {
		return s.sessionClient.CloseSession(ctx, sessionID)
	})
	if err != nil {
		return nil, err
	}
	return &domain.ModerationResult{SessionEnded: true, Log: *entry}, nil
}

func (s *service) ViewLogs(ctx context.Context, action string) ([]domain.AdminLog, error) {
	a := domain.AdminAction(strings.TrimSpace(action))
	if a != "" && !a.Valid() {
		return nil, fmt.Errorf("%w: Unknown action %q.", domain.ErrInvalid, action)
	}
	return s.logs.List(ctx, a, DefaultLogLimit)
}

// requireLive turns "room is not live" into ErrStateChanged, so the admin is
// told the room changed under them and refreshes (UC-10 E-3).
func (s *service) requireLive(ctx context.Context, sessionID string) (*domain.SessionOverview, error) {
	room, err := s.sessions.GetActiveSession(ctx, sessionID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("%w: This room is no longer active.", domain.ErrStateChanged)
	}
	return room, err
}

func newLog(adminID string, action domain.AdminAction, targetID, reason string) (*domain.AdminLog, error) {
	// admin_logs.admin_id is a UUID column; the ID comes from the JWT sub.
	if !isUUID(adminID) {
		return nil, fmt.Errorf("%w: Your admin id is not valid. Sign in again.", domain.ErrInvalid)
	}
	entry := &domain.AdminLog{AdminID: adminID, Action: action, TargetID: targetID}
	if r := strings.TrimSpace(reason); r != "" {
		// Counted in characters (runes), as the API docs say: a byte count would
		// reject a 200-character Thai reason that the form allows.
		if utf8.RuneCountInString(r) > domain.MaxReasonLength {
			return nil, fmt.Errorf("%w: The reason must be at most %d characters.", domain.ErrInvalid, domain.MaxReasonLength)
		}
		entry.Reason = &r
	}
	return entry, nil
}

// isUUID matches the UUID columns of admin_logs and Study Session's own check.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
				return false
			}
		}
	}
	return true
}
