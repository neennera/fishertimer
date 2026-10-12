package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("admin: resource not found")
	ErrInvalid  = errors.New("admin: invalid input")
	ErrUpstream = errors.New("admin: study session service unavailable")
	// ErrStateChanged is returned when a moderation target is no longer in
	// the state the admin saw: the room has ended or the participant has
	// already left (UC-04 E-4, UC-10 E-3). Nothing was changed.
	ErrStateChanged = errors.New("admin: room or participant state has changed")
)

// MaxReasonLength caps the optional free-text reason on a moderation action.
const MaxReasonLength = 500

// AdminAction is the action type stored in admin_logs.action.
type AdminAction string

const (
	ActionKickUser          AdminAction = "KICK_USER"
	ActionForceCloseSession AdminAction = "FORCE_CLOSE_SESSION"
)

func (a AdminAction) Valid() bool {
	return a == ActionKickUser || a == ActionForceCloseSession
}

// SessionOverview is a live room as Study Session reports it.
type SessionOverview struct {
	SessionID        string     `json:"session_id"`
	Title            string     `json:"title"`
	HostID           string     `json:"host_id"`
	HostName         *string    `json:"host_name"`
	IsActive         bool       `json:"is_active"`
	MaxParticipants  int        `json:"max_participants"`
	ParticipantCount int        `json:"participant_count"`
	CreatedAt        time.Time  `json:"created_at"`
	EndedAt          *time.Time `json:"ended_at"`
}

type SessionParticipantOverview struct {
	SessionID   string     `json:"session_id"`
	UserID      string     `json:"user_id"`
	DisplayName *string    `json:"display_name"`
	IsHost      bool       `json:"is_host"`
	JoinedAt    time.Time  `json:"joined_at"`
	LeftAt      *time.Time `json:"left_at"`
}

// SessionDetail is a live room together with its current participants
// (ViewSessionDetails).
type SessionDetail struct {
	SessionOverview
	Participants []SessionParticipantOverview `json:"participants"`
}

// AdminLog is one row of admin_db.admin_logs. TargetID is the user ID for
// KICK_USER and the session ID for FORCE_CLOSE_SESSION.
type AdminLog struct {
	LogID     string      `json:"log_id"`
	AdminID   string      `json:"admin_id"`
	Action    AdminAction `json:"action"`
	TargetID  string      `json:"target_id"`
	Reason    *string     `json:"reason"`
	CreatedAt time.Time   `json:"created_at"`
}

// ModerationResult describes what a kick or close did.
type ModerationResult struct {
	// SessionEnded is true when the room is now ENDED: always for a close,
	// and for a kick that removed the last participant.
	SessionEnded bool     `json:"session_ended"`
	Log          AdminLog `json:"log"`
}
