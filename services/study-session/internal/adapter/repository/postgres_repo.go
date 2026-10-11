package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// oneActiveRoomIndex is the partial unique index that keeps a user in at
// most one room (see database/schemas/002_phase2_rooms_and_outbox.sql).
const oneActiveRoomIndex = "uq_session_participants_one_active_room"

// PostgresRepository persists rooms, participants and the event outbox in
// session_db. Every command runs in one transaction that also writes its
// events to event_outbox, so a room change and its events commit together.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const sessionColumns = `session_id::text, title, host_id::text, max_participants,
	participant_count, status, created_at, ended_at, COALESCE(end_reason, '')`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSession(row rowScanner) (*domain.StudySession, error) {
	var s domain.StudySession
	var status string
	var endedAt sql.NullTime
	err := row.Scan(&s.ID, &s.Name, &s.CreatorID, &s.ParticipantLimit,
		&s.ParticipantCount, &status, &s.CreatedAt, &endedAt, &s.EndReason)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.Status = domain.SessionStatus(status)
	if endedAt.Valid {
		t := endedAt.Time
		s.EndedAt = &t
	}
	return &s, nil
}

const participantColumns = `session_id::text, user_id::text, display_name, joined_at,
	last_seen_at, left_at, COALESCE(leave_reason, '')`

func scanParticipant(row rowScanner) (*domain.Participant, error) {
	var p domain.Participant
	var leftAt sql.NullTime
	err := row.Scan(&p.SessionID, &p.UserID, &p.DisplayName, &p.JoinedAt,
		&p.LastSeenAt, &leftAt, &p.LeaveReason)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if leftAt.Valid {
		t := leftAt.Time
		p.LeftAt = &t
	}
	return &p, nil
}

func isOneActiveRoomViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == oneActiveRoomIndex
}

func enqueue(ctx context.Context, tx *sql.Tx, ev domain.OutboxEvent) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO event_outbox (event_id, routing_key, payload, created_at)
		VALUES ($1, $2, $3, $4)`,
		ev.EventID, ev.RoutingKey, ev.Payload, ev.CreatedAt)
	return err
}

func (r *PostgresRepository) CreateSession(ctx context.Context, s *domain.StudySession, creator *domain.Participant, joined domain.OutboxEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO study_sessions (session_id, title, host_id, max_participants, participant_count, status, created_at)
		VALUES ($1, $2, $3, $4, 1, 'ACTIVE', $5)`,
		s.ID, s.Name, s.CreatorID, s.ParticipantLimit, s.CreatedAt,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO session_participants (session_id, user_id, display_name, joined_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $4)`,
		s.ID, creator.UserID, creator.DisplayName, creator.JoinedAt,
	); err != nil {
		if isOneActiveRoomViolation(err) {
			return domain.ErrAlreadyInSession
		}
		return err
	}
	if err := enqueue(ctx, tx, joined); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresRepository) JoinSession(ctx context.Context, p *domain.Participant, joined domain.OutboxEvent) (*domain.StudySession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Already in a room? The same room is a harmless re-join (a reload, or
	// a reconnect after a dropped connection - UC-02 E-5): no new row.
	var currentRoom string
	err = tx.QueryRowContext(ctx, `
		SELECT session_id::text FROM session_participants
		WHERE user_id = $1 AND left_at IS NULL`, p.UserID,
	).Scan(&currentRoom)
	switch {
	case err == nil && currentRoom == p.SessionID:
		sess, err := scanSession(tx.QueryRowContext(ctx,
			`SELECT `+sessionColumns+` FROM study_sessions WHERE session_id = $1`, p.SessionID))
		if err != nil {
			return nil, err
		}
		return sess, tx.Commit()
	case err == nil:
		return nil, domain.ErrAlreadyInSession
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}

	// Take a seat atomically: the row only updates while the room is ACTIVE
	// and below its limit, so two users racing for the last seat cannot both
	// get it (UC-02 E-2). The UPDATE also locks the room row until commit.
	sess, err := scanSession(tx.QueryRowContext(ctx, `
		UPDATE study_sessions
		SET participant_count = participant_count + 1
		WHERE session_id = $1 AND status = 'ACTIVE' AND participant_count < max_participants
		RETURNING `+sessionColumns, p.SessionID))
	if errors.Is(err, domain.ErrNotFound) {
		return nil, r.whyNoSeat(ctx, tx, p.SessionID)
	}
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO session_participants (session_id, user_id, display_name, joined_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $4)`,
		p.SessionID, p.UserID, p.DisplayName, p.JoinedAt,
	); err != nil {
		if isOneActiveRoomViolation(err) {
			// Lost a race with a join to another room; the rollback also
			// gives the seat back.
			return nil, domain.ErrAlreadyInSession
		}
		return nil, err
	}
	if err := enqueue(ctx, tx, joined); err != nil {
		return nil, err
	}
	return sess, tx.Commit()
}

// whyNoSeat explains a failed seat update: the room is missing, ended or full.
func (r *PostgresRepository) whyNoSeat(ctx context.Context, tx *sql.Tx, sessionID string) error {
	var status string
	err := tx.QueryRowContext(ctx, `SELECT status FROM study_sessions WHERE session_id = $1`, sessionID).Scan(&status)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.ErrNotFound
	case err != nil:
		return err
	case status != string(domain.StatusActive):
		return domain.ErrSessionEnded
	default:
		return domain.ErrSessionFull
	}
}

func (r *PostgresRepository) LeaveSession(ctx context.Context, sessionID, userID, reason string, now time.Time, left, ended domain.OutboxEvent) (*domain.LeaveResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock the room first, so the last leave cannot interleave with a join
	// or an EndSession on the same room.
	if _, err := scanSession(tx.QueryRowContext(ctx,
		`SELECT `+sessionColumns+` FROM study_sessions WHERE session_id = $1 FOR UPDATE`, sessionID)); err != nil {
		return nil, err
	}

	p, err := scanParticipant(tx.QueryRowContext(ctx, `
		UPDATE session_participants
		SET left_at = $3, leave_reason = $4
		WHERE session_id = $1 AND user_id = $2 AND left_at IS NULL
		RETURNING `+participantColumns, sessionID, userID, now, reason))
	if errors.Is(err, domain.ErrNotFound) {
		// Already left (or never joined): nothing to do (UC-04 E-4).
		return &domain.LeaveResult{}, tx.Commit()
	}
	if err != nil {
		return nil, err
	}

	var remaining int
	if err := tx.QueryRowContext(ctx, `
		UPDATE study_sessions SET participant_count = participant_count - 1
		WHERE session_id = $1 RETURNING participant_count`, sessionID,
	).Scan(&remaining); err != nil {
		return nil, err
	}
	if err := enqueue(ctx, tx, left); err != nil {
		return nil, err
	}

	res := &domain.LeaveResult{Left: true, Participant: p}
	if remaining == 0 {
		// Last one out ends the room (UC-03 S-2).
		if _, err := tx.ExecContext(ctx, `
			UPDATE study_sessions SET status = 'ENDED', ended_at = $2, end_reason = 'EMPTY'
			WHERE session_id = $1 AND status = 'ACTIVE'`, sessionID, now,
		); err != nil {
			return nil, err
		}
		if err := enqueue(ctx, tx, ended); err != nil {
			return nil, err
		}
		res.SessionEnded = true
	}
	return res, tx.Commit()
}

func (r *PostgresRepository) EndSession(ctx context.Context, sessionID, reason string, now time.Time, ended domain.OutboxEvent) (*domain.StudySession, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	sess, err := scanSession(tx.QueryRowContext(ctx,
		`SELECT `+sessionColumns+` FROM study_sessions WHERE session_id = $1 FOR UPDATE`, sessionID))
	if err != nil {
		return nil, false, err
	}
	if sess.Status == domain.StatusEnded {
		return sess, false, tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE session_participants SET left_at = $2, leave_reason = 'SESSION_ENDED'
		WHERE session_id = $1 AND left_at IS NULL`, sessionID, now,
	); err != nil {
		return nil, false, err
	}
	sess, err = scanSession(tx.QueryRowContext(ctx, `
		UPDATE study_sessions
		SET status = 'ENDED', ended_at = $2, end_reason = $3, participant_count = 0
		WHERE session_id = $1
		RETURNING `+sessionColumns, sessionID, now, reason))
	if err != nil {
		return nil, false, err
	}
	// One session.ended finalizes every timer in the room; Study Timer does
	// not need a left event per participant.
	if err := enqueue(ctx, tx, ended); err != nil {
		return nil, false, err
	}
	return sess, true, tx.Commit()
}

func (r *PostgresRepository) GetSession(ctx context.Context, id string) (*domain.StudySession, error) {
	return scanSession(r.db.QueryRowContext(ctx,
		`SELECT `+sessionColumns+` FROM study_sessions WHERE session_id = $1`, id))
}

func (r *PostgresRepository) ListActiveSessions(ctx context.Context) ([]domain.StudySession, error) {
	return r.listSessions(ctx, `SELECT `+sessionColumns+` FROM study_sessions
		WHERE status = 'ACTIVE' ORDER BY created_at DESC`)
}

func (r *PostgresRepository) ListSessionsCreatedBefore(ctx context.Context, cutoff time.Time) ([]domain.StudySession, error) {
	return r.listSessions(ctx, `SELECT `+sessionColumns+` FROM study_sessions
		WHERE status = 'ACTIVE' AND created_at < $1`, cutoff)
}

func (r *PostgresRepository) listSessions(ctx context.Context, query string, args ...any) ([]domain.StudySession, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := []domain.StudySession{}
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, *s)
	}
	return sessions, rows.Err()
}

func (r *PostgresRepository) GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error) {
	return r.listParticipants(ctx, `SELECT `+participantColumns+` FROM session_participants
		WHERE session_id = $1 AND left_at IS NULL ORDER BY joined_at`, sessionID)
}

func (r *PostgresRepository) ListStaleParticipants(ctx context.Context, cutoff time.Time) ([]domain.Participant, error) {
	return r.listParticipants(ctx, `SELECT `+participantColumns+` FROM session_participants
		WHERE left_at IS NULL AND last_seen_at < $1`, cutoff)
}

func (r *PostgresRepository) listParticipants(ctx context.Context, query string, args ...any) ([]domain.Participant, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	participants := []domain.Participant{}
	for rows.Next() {
		p, err := scanParticipant(rows)
		if err != nil {
			return nil, err
		}
		participants = append(participants, *p)
	}
	return participants, rows.Err()
}

func (r *PostgresRepository) GetActiveParticipation(ctx context.Context, userID string) (*domain.Participant, error) {
	return scanParticipant(r.db.QueryRowContext(ctx, `SELECT `+participantColumns+`
		FROM session_participants WHERE user_id = $1 AND left_at IS NULL`, userID))
}

func (r *PostgresRepository) GetLastParticipation(ctx context.Context, sessionID, userID string) (*domain.Participant, error) {
	return scanParticipant(r.db.QueryRowContext(ctx, `SELECT `+participantColumns+`
		FROM session_participants WHERE session_id = $1 AND user_id = $2
		ORDER BY joined_at DESC LIMIT 1`, sessionID, userID))
}

func (r *PostgresRepository) TouchParticipant(ctx context.Context, sessionID, userID string, now time.Time) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE session_participants SET last_seen_at = $3
		WHERE session_id = $1 AND user_id = $2 AND left_at IS NULL`, sessionID, userID, now)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *PostgresRepository) PendingEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id::text, routing_key, payload, created_at, attempts
		FROM event_outbox WHERE published_at IS NULL
		ORDER BY id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pending []domain.OutboxEvent
	for rows.Next() {
		var ev domain.OutboxEvent
		if err := rows.Scan(&ev.ID, &ev.EventID, &ev.RoutingKey, &ev.Payload, &ev.CreatedAt, &ev.Attempts); err != nil {
			return nil, err
		}
		pending = append(pending, ev)
	}
	return pending, rows.Err()
}

func (r *PostgresRepository) MarkEventPublished(ctx context.Context, id int64, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE event_outbox SET published_at = $2, attempts = attempts + 1, last_error = NULL
		WHERE id = $1`, id, now)
	return err
}

func (r *PostgresRepository) MarkEventFailed(ctx context.Context, id int64, cause string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE event_outbox SET attempts = attempts + 1, last_error = $2
		WHERE id = $1`, id, cause)
	return err
}
