package usecase

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// LeaveReasonSessionEnded marks participations closed because their room
// ended, as opposed to the user leaving on their own.
const LeaveReasonSessionEnded = "SESSION_ENDED"

type Usecase interface {
	// CreateSession opens a room with the creator as its first participant (UC-01).
	CreateSession(ctx context.Context, name, creatorID, creatorName string, limit int) (*domain.StudySession, error)
	// JoinSession adds the user to an active room with free capacity (UC-02).
	JoinSession(ctx context.Context, sessionID, userID, displayName string) (*domain.StudySession, error)
	// LeaveSession closes the user's participation (UC-03, UC-04 S-2). The
	// last one out ends the room. An empty reason means LEFT.
	LeaveSession(ctx context.Context, sessionID, userID, reason string) (*domain.LeaveResult, error)
	// EndSession closes a room and everyone in it (UC-04 S-3). An empty
	// reason means ADMIN_CLOSED.
	EndSession(ctx context.Context, sessionID, reason string) (*domain.StudySession, error)
	ListActiveSession(ctx context.Context) ([]domain.StudySession, error)
	GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error)
	GetSession(ctx context.Context, sessionID string) (*domain.StudySession, error)
	// GetMySession returns the room the user is in, or (nil, nil, nil).
	GetMySession(ctx context.Context, userID string) (*domain.StudySession, *domain.Participant, error)
	// Heartbeat records presence. When the user is no longer in the room it
	// reports active = false and, when known, why.
	Heartbeat(ctx context.Context, sessionID, userID string) (active bool, reason string, err error)
}

type service struct {
	repo domain.Repository
	now  func() time.Time
}

// Option customises the usecase service.
type Option func(*service)

// WithClock replaces the wall clock, so tests can control time.
func WithClock(now func() time.Time) Option {
	return func(s *service) { s.now = now }
}

func New(repo domain.Repository, opts ...Option) Usecase {
	s := &service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) CreateSession(ctx context.Context, name, creatorID, creatorName string, limit int) (*domain.StudySession, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > domain.MaxNameLength {
		return nil, fmt.Errorf("%w: room name must be 1-%d characters", domain.ErrInvalid, domain.MaxNameLength)
	}
	if limit < domain.MinParticipantLimit || limit > domain.MaxParticipantLimit {
		return nil, fmt.Errorf("%w: participant limit must be %d-%d", domain.ErrInvalid, domain.MinParticipantLimit, domain.MaxParticipantLimit)
	}
	if !isUUID(creatorID) {
		return nil, fmt.Errorf("%w: creator id must be a UUID", domain.ErrInvalid)
	}

	now := s.now()
	sess := &domain.StudySession{
		ID:               newUUID(),
		Name:             name,
		CreatorID:        creatorID,
		ParticipantLimit: limit,
		ParticipantCount: 1,
		Status:           domain.StatusActive,
		CreatedAt:        now,
	}
	creator := &domain.Participant{
		SessionID:   sess.ID,
		UserID:      creatorID,
		DisplayName: cleanDisplayName(creatorName),
		JoinedAt:    now,
		LastSeenAt:  now,
	}
	if err := s.repo.CreateSession(ctx, sess, creator, joinedEvent(sess.ID, creatorID, now)); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *service) JoinSession(ctx context.Context, sessionID, userID, displayName string) (*domain.StudySession, error) {
	if !isUUID(sessionID) || !isUUID(userID) {
		return nil, fmt.Errorf("%w: session id and user id must be UUIDs", domain.ErrInvalid)
	}
	now := s.now()
	p := &domain.Participant{
		SessionID:   sessionID,
		UserID:      userID,
		DisplayName: cleanDisplayName(displayName),
		JoinedAt:    now,
		LastSeenAt:  now,
	}
	return s.repo.JoinSession(ctx, p, joinedEvent(sessionID, userID, now))
}

var leaveReasons = map[string]bool{
	events.ReasonLeft:              true,
	events.ReasonKicked:            true,
	events.ReasonDisconnectTimeout: true,
	events.ReasonIdleTimeout:       true,
}

func (s *service) LeaveSession(ctx context.Context, sessionID, userID, reason string) (*domain.LeaveResult, error) {
	if !isUUID(sessionID) || !isUUID(userID) {
		return nil, fmt.Errorf("%w: session id and user id must be UUIDs", domain.ErrInvalid)
	}
	if reason == "" {
		reason = events.ReasonLeft
	}
	if !leaveReasons[reason] {
		return nil, fmt.Errorf("%w: unknown leave reason %q", domain.ErrInvalid, reason)
	}

	now := s.now()
	left := mustEvent(events.RoutingKeyLeft, events.ParticipantLeft{
		EventID:    newUUID(),
		OccurredAt: now,
		SessionID:  sessionID,
		UserID:     userID,
		Reason:     reason,
	})
	return s.repo.LeaveSession(ctx, sessionID, userID, reason, now, left, endedEvent(sessionID, events.ReasonEmpty, now))
}

var endReasons = map[string]bool{
	events.ReasonEmpty:       true,
	events.ReasonAdminClosed: true,
	events.ReasonTimeout24H:  true,
}

func (s *service) EndSession(ctx context.Context, sessionID, reason string) (*domain.StudySession, error) {
	if !isUUID(sessionID) {
		return nil, fmt.Errorf("%w: session id must be a UUID", domain.ErrInvalid)
	}
	if reason == "" {
		reason = events.ReasonAdminClosed
	}
	if !endReasons[reason] {
		return nil, fmt.Errorf("%w: unknown end reason %q", domain.ErrInvalid, reason)
	}
	now := s.now()
	sess, _, err := s.repo.EndSession(ctx, sessionID, reason, now, endedEvent(sessionID, reason, now))
	return sess, err
}

func (s *service) ListActiveSession(ctx context.Context) ([]domain.StudySession, error) {
	return s.repo.ListActiveSessions(ctx)
}

func (s *service) GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error) {
	if !isUUID(sessionID) {
		return nil, fmt.Errorf("%w: session id must be a UUID", domain.ErrInvalid)
	}
	return s.repo.GetParticipants(ctx, sessionID)
}

func (s *service) GetSession(ctx context.Context, sessionID string) (*domain.StudySession, error) {
	if !isUUID(sessionID) {
		return nil, fmt.Errorf("%w: session id must be a UUID", domain.ErrInvalid)
	}
	return s.repo.GetSession(ctx, sessionID)
}

func (s *service) GetMySession(ctx context.Context, userID string) (*domain.StudySession, *domain.Participant, error) {
	if !isUUID(userID) {
		return nil, nil, fmt.Errorf("%w: user id must be a UUID", domain.ErrInvalid)
	}
	p, err := s.repo.GetActiveParticipation(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	sess, err := s.repo.GetSession(ctx, p.SessionID)
	if err != nil {
		return nil, nil, err
	}
	return sess, p, nil
}

func (s *service) Heartbeat(ctx context.Context, sessionID, userID string) (bool, string, error) {
	if !isUUID(sessionID) || !isUUID(userID) {
		return false, "", fmt.Errorf("%w: session id and user id must be UUIDs", domain.ErrInvalid)
	}
	active, err := s.repo.TouchParticipant(ctx, sessionID, userID, s.now())
	if err != nil || active {
		return active, "", err
	}
	last, err := s.repo.GetLastParticipation(ctx, sessionID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return false, last.LeaveReason, nil
}

func joinedEvent(sessionID, userID string, now time.Time) domain.OutboxEvent {
	return mustEvent(events.RoutingKeyJoined, events.ParticipantJoined{
		EventID:    newUUID(),
		OccurredAt: now,
		SessionID:  sessionID,
		UserID:     userID,
	})
}

func endedEvent(sessionID, reason string, now time.Time) domain.OutboxEvent {
	return mustEvent(events.RoutingKeyEnded, events.SessionEnded{
		EventID:    newUUID(),
		OccurredAt: now,
		SessionID:  sessionID,
		Reason:     reason,
	})
}

// mustEvent wraps one of the pkg/events payloads for the outbox. The
// payloads are plain structs, so marshalling cannot fail.
func mustEvent(routingKey string, payload any) domain.OutboxEvent {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("study-session: marshal %s event: %v", routingKey, err))
	}
	var head struct {
		EventID    string    `json:"event_id"`
		OccurredAt time.Time `json:"occurred_at"`
	}
	_ = json.Unmarshal(body, &head)
	return domain.OutboxEvent{
		EventID:    head.EventID,
		RoutingKey: routingKey,
		Payload:    body,
		CreatedAt:  head.OccurredAt,
	}
}

// cleanDisplayName trims the name shown in the room roster; the account
// service already validated it, this only guards the column width.
func cleanDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) > 100 {
		name = string([]rune(name)[:100])
	}
	return name
}

// newUUID returns a random (version 4) UUID.
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("study-session: read random bytes: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// isUUID reports whether s is a canonical 8-4-4-4-12 hex UUID. Ids are
// UUID columns in session_db and timer_db, so anything else is rejected up
// front as invalid input rather than surfacing as a database error.
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
