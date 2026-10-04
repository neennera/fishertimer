package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// singleActiveIndex keeps one running or paused cycle per timer
// (database/schemas/003_phase2_timer_and_events.sql, UC-05 E-4).
const singleActiveIndex = "idx_cycles_single_active"

// PostgresRepository persists timers in timer_db: timers (one per
// participant per room, OPEN / FINALIZED), cycles (every work and rest
// period with its timestamps and reward delivery), timer_settings and
// processed_events. The Phase 1 tables timer_sessions / timer_cycles are no
// longer written.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const cycleColumns = `c.cycle_id::text, c.timer_id::text, c.type, c.status, c.duration_sec,
	c.started_at, c.paused_at, c.paused_total_sec, c.ended_at, c.reward_status`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCycle(row rowScanner, extra ...any) (*domain.Cycle, error) {
	var c domain.Cycle
	var typ, status, reward string
	var pausedAt, endedAt sql.NullTime
	dest := append([]any{&c.CycleID, &c.TimerID, &typ, &status, &c.DurationSec,
		&c.StartedAt, &pausedAt, &c.PausedTotalSec, &endedAt, &reward}, extra...)
	if err := row.Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	c.Type, c.Status, c.RewardStatus = domain.TimerPhase(typ), domain.CycleStatus(status), domain.RewardStatus(reward)
	if pausedAt.Valid {
		t := pausedAt.Time
		c.PausedAt = &t
	}
	if endedAt.Valid {
		t := endedAt.Time
		c.EndedAt = &t
	}
	return &c, nil
}

func optionalCycle(c *domain.Cycle, err error) (*domain.Cycle, error) {
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	}
	return c, err
}

func (r *PostgresRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.Timer, error) {
	t := &domain.Timer{SessionID: sessionID, UserID: userID}
	var status string
	var finalizedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT timer_id::text, status, opened_at, finalized_at
		FROM timers WHERE session_id = $1 AND user_id = $2`, sessionID, userID,
	).Scan(&t.TimerID, &status, &t.OpenedAt, &finalizedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Status = domain.SessionTimerStatus(status)
	if finalizedAt.Valid {
		at := finalizedAt.Time
		t.FinalizedAt = &at
	}

	if t.Settings, err = r.GetSettings(ctx, userID); err != nil {
		return nil, err
	}
	if t.Active, err = optionalCycle(scanCycle(r.db.QueryRowContext(ctx, `
		SELECT `+cycleColumns+` FROM cycles c
		WHERE c.timer_id = $1 AND c.status IN ('RUNNING', 'PAUSED')`, t.TimerID))); err != nil {
		return nil, err
	}
	if t.Last, err = optionalCycle(scanCycle(r.db.QueryRowContext(ctx, `
		SELECT `+cycleColumns+` FROM cycles c
		WHERE c.timer_id = $1 AND c.status NOT IN ('RUNNING', 'PAUSED')
		ORDER BY c.ended_at DESC NULLS LAST, c.started_at DESC LIMIT 1`, t.TimerID))); err != nil {
		return nil, err
	}

	// Work done during this stay: cycles started since the timer was
	// (re)opened, so a rejoin starts the room summary from zero.
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(duration_sec), 0) FROM cycles
		WHERE timer_id = $1 AND type = 'WORK' AND status = 'COMPLETED' AND started_at >= $2`,
		t.TimerID, t.OpenedAt,
	).Scan(&t.CompletedWork, &t.FocusSeconds); err != nil {
		return nil, err
	}
	if t.LastCompletedWork, err = optionalCycle(scanCycle(r.db.QueryRowContext(ctx, `
		SELECT `+cycleColumns+` FROM cycles c
		WHERE c.timer_id = $1 AND c.type = 'WORK' AND c.status = 'COMPLETED' AND c.started_at >= $2
		ORDER BY c.ended_at DESC LIMIT 1`, t.TimerID, t.OpenedAt))); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *PostgresRepository) ListRoomTimers(ctx context.Context, sessionID string) ([]*domain.Timer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id::text FROM timers WHERE session_id = $1 AND status = 'OPEN'
		ORDER BY opened_at`, sessionID)
	if err != nil {
		return nil, err
	}
	var users []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return nil, err
		}
		users = append(users, u)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	timers := make([]*domain.Timer, 0, len(users))
	for _, u := range users {
		t, err := r.GetTimer(ctx, sessionID, u)
		if errors.Is(err, domain.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		timers = append(timers, t)
	}
	return timers, nil
}

func (r *PostgresRepository) EnsureTimer(ctx context.Context, sessionID, userID string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO timers (session_id, user_id, status, opened_at)
		VALUES ($1, $2, 'OPEN', $3)
		ON CONFLICT (session_id, user_id) DO NOTHING`, sessionID, userID, now)
	return err
}

func (r *PostgresRepository) InsertCycle(ctx context.Context, c *domain.Cycle) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO cycles (timer_id, type, status, duration_sec, started_at, paused_at, paused_total_sec, ended_at, reward_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING cycle_id::text`,
		c.TimerID, c.Type, c.Status, c.DurationSec, c.StartedAt, c.PausedAt, c.PausedTotalSec, c.EndedAt, c.RewardStatus,
	).Scan(&c.CycleID)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == singleActiveIndex {
		return domain.ErrInvalidState
	}
	return err
}

func (r *PostgresRepository) UpdateCycle(ctx context.Context, c *domain.Cycle, from ...domain.CycleStatus) error {
	statuses := make([]string, len(from))
	for i, s := range from {
		statuses[i] = string(s)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE cycles
		SET status = $2, started_at = $3, paused_at = $4, paused_total_sec = $5,
		    ended_at = $6, reward_status = $7
		WHERE cycle_id = $1 AND status = ANY($8)`,
		c.CycleID, c.Status, c.StartedAt, c.PausedAt, c.PausedTotalSec, c.EndedAt, c.RewardStatus, pq.Array(statuses))
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *PostgresRepository) GetSettings(ctx context.Context, userID string) (domain.Settings, error) {
	s := domain.DefaultSettings
	err := r.db.QueryRowContext(ctx, `
		SELECT focus_duration / 60, short_break_duration / 60
		FROM timer_settings WHERE user_id = $1`, userID,
	).Scan(&s.WorkMinutes, &s.RestMinutes)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultSettings, nil
	}
	return s, err
}

func (r *PostgresRepository) SaveSettings(ctx context.Context, userID string, s domain.Settings) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO timer_settings (user_id, focus_duration, short_break_duration)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET focus_duration = EXCLUDED.focus_duration,
		    short_break_duration = EXCLUDED.short_break_duration,
		    updated_at = CURRENT_TIMESTAMP`,
		userID, s.WorkMinutes*60, s.RestMinutes*60)
	return err
}

func (r *PostgresRepository) listRefs(ctx context.Context, query string, args ...any) ([]domain.CycleRef, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []domain.CycleRef
	for rows.Next() {
		var ref domain.CycleRef
		c, err := scanCycle(rows, &ref.SessionID, &ref.UserID)
		if err != nil {
			return nil, err
		}
		ref.Cycle = *c
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

const refFrom = ` FROM cycles c JOIN timers t ON t.timer_id = c.timer_id `
const refColumns = cycleColumns + `, t.session_id::text, t.user_id::text`

func (r *PostgresRepository) DueCycles(ctx context.Context, now time.Time, limit int) ([]domain.CycleRef, error) {
	return r.listRefs(ctx, `SELECT `+refColumns+refFrom+`
		WHERE c.status = 'RUNNING' AND t.status = 'OPEN'
		  AND c.started_at + make_interval(secs => c.duration_sec + c.paused_total_sec) <= $1
		ORDER BY c.started_at LIMIT $2`, now, limit)
}

func (r *PostgresRepository) StalePausedCycles(ctx context.Context, cutoff time.Time, limit int) ([]domain.CycleRef, error) {
	return r.listRefs(ctx, `SELECT `+refColumns+refFrom+`
		WHERE c.status = 'PAUSED' AND t.status = 'OPEN' AND c.paused_at < $1
		ORDER BY c.paused_at LIMIT $2`, cutoff, limit)
}

func (r *PostgresRepository) PendingRewards(ctx context.Context, limit int) ([]domain.CycleRef, error) {
	return r.listRefs(ctx, `SELECT `+refColumns+refFrom+`
		WHERE c.type = 'WORK' AND c.status = 'COMPLETED' AND c.reward_status = 'PENDING'
		ORDER BY c.ended_at LIMIT $1`, limit)
}

func (r *PostgresRepository) MarkRewardSent(ctx context.Context, cycleID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE cycles SET reward_status = 'SENT'
		WHERE cycle_id = $1 AND reward_status = 'PENDING'`, cycleID)
	return err
}

func (r *PostgresRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	h := &domain.TimerHistory{UserID: userID}

	var lastJoined sql.NullTime
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(opened_at) FROM timers WHERE user_id = $1`, userID,
	).Scan(&h.SessionsJoined, &lastJoined); err != nil {
		return nil, err
	}
	if lastJoined.Valid {
		h.LastActive = lastJoined.Time
	}

	var focusSeconds int
	var lastCycle sql.NullTime
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(c.duration_sec), 0), MAX(c.ended_at)`+refFrom+`
		WHERE t.user_id = $1 AND c.type = 'WORK' AND c.status = 'COMPLETED'`, userID,
	).Scan(&h.CyclesCompleted, &focusSeconds, &lastCycle); err != nil {
		return nil, err
	}
	h.TotalFocusMinutes = focusSeconds / 60
	if lastCycle.Valid && lastCycle.Time.After(h.LastActive) {
		h.LastActive = lastCycle.Time
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT (c.ended_at AT TIME ZONE 'UTC')::date AS day, SUM(c.duration_sec) / 60`+refFrom+`
		WHERE t.user_id = $1 AND c.type = 'WORK' AND c.status = 'COMPLETED'
		  AND c.ended_at >= CURRENT_DATE - INTERVAL '29 days'
		GROUP BY day`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byDay := make(map[string]int)
	for rows.Next() {
		var day time.Time
		var minutes int
		if err := rows.Scan(&day, &minutes); err != nil {
			return nil, err
		}
		byDay[day.Format("2006-01-02")] = minutes
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	h.DailyFocusMinutes = last30Days(byDay, time.Now().UTC())
	return h, nil
}

// last30Days lists the 30 days ending today, oldest first, zero days included.
func last30Days(minutesByDay map[string]int, now time.Time) []domain.DailyFocus {
	today := now.Truncate(24 * time.Hour)
	daily := make([]domain.DailyFocus, 30)
	for i := range 30 {
		date := today.AddDate(0, 0, -(29 - i)).Format("2006-01-02")
		daily[i] = domain.DailyFocus{Date: date, FocusMinutes: minutesByDay[date]}
	}
	return daily
}

// claimEvent records eventID inside tx. It reports false when the event was
// processed before (at-least-once delivery), in which case the caller does
// nothing.
func claimEvent(ctx context.Context, tx *sql.Tx, eventID, eventType string) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO processed_events (event_id, event_type) VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING`, eventID, eventType)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *PostgresRepository) OpenParticipantTimer(ctx context.Context, sessionID, userID, eventID string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if fresh, err := claimEvent(ctx, tx, eventID, "session.participant.joined"); err != nil || !fresh {
		return err
	}
	// New stay: open the timer, or reopen it if the user left this room
	// before. An already open timer (a Start that beat this event) is kept.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO timers (session_id, user_id, status, opened_at)
		VALUES ($1, $2, 'OPEN', $3)
		ON CONFLICT (session_id, user_id) DO UPDATE
		SET status = 'OPEN', opened_at = EXCLUDED.opened_at, finalized_at = NULL
		WHERE timers.status = 'FINALIZED'`, sessionID, userID, now); err != nil {
		return err
	}
	return tx.Commit()
}

// finalize closes the timers matched by where (on the timers table, $1 is
// the session id, $2 optionally the user id). A work cycle whose time had
// already run out is completed first, so its reward is not lost to a leave
// that raced the sweeper; anything still in progress is discarded.
func (r *PostgresRepository) finalize(ctx context.Context, eventID, eventType, where string, args ...any) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if fresh, err := claimEvent(ctx, tx, eventID, eventType); err != nil || !fresh {
		return err
	}
	now := time.Now().UTC()
	n := len(args)

	if _, err := tx.ExecContext(ctx, `
		UPDATE cycles c
		SET status = 'COMPLETED',
		    ended_at = c.started_at + make_interval(secs => c.duration_sec + c.paused_total_sec),
		    reward_status = CASE WHEN c.type = 'WORK' THEN 'PENDING' ELSE c.reward_status END
		WHERE c.status = 'RUNNING'
		  AND c.started_at + make_interval(secs => c.duration_sec + c.paused_total_sec) <= $`+strconv.Itoa(n+1)+`
		  AND c.timer_id IN (SELECT timer_id FROM timers WHERE `+where+`)`, append(args, now)...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE cycles
		SET status = 'DISCARDED', ended_at = COALESCE(paused_at, $`+strconv.Itoa(n+1)+`), paused_at = NULL
		WHERE status IN ('RUNNING', 'PAUSED')
		  AND timer_id IN (SELECT timer_id FROM timers WHERE `+where+`)`, append(args, now)...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE timers SET status = 'FINALIZED', finalized_at = $`+strconv.Itoa(n+1)+`
		WHERE status = 'OPEN' AND `+where, append(args, now)...); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresRepository) FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, eventType string) error {
	return r.finalize(ctx, eventID, eventType, `session_id = $1 AND user_id = $2`, sessionID, userID)
}

func (r *PostgresRepository) FinalizeSessionTimers(ctx context.Context, sessionID, eventID, eventType string) error {
	return r.finalize(ctx, eventID, eventType, `session_id = $1`, sessionID)
}

func (r *PostgresRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM processed_events WHERE event_id = $1`, eventID).Scan(&count)
	return count > 0, err
}
