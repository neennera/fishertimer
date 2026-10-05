package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// PostgresRepository persists timer state in timer_db (timer_settings,
// timer_sessions, timer_cycles) - see database/schemas/.
//
// The schema's timer_sessions.status enum (FOCUS/SHORT_BREAK/LONG_BREAK/
// PAUSED/COMPLETED/STOPPED) captures the *current running phase* while a
// timer is active, collapsing to PAUSED/STOPPED/COMPLETED otherwise. The
// live progress the domain derives remaining time from (phase,
// running_since, elapsed_ms) is stored alongside it, so a reload restores
// the timer exactly.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t := domain.NewTimer(sessionID, userID)

	if err := r.db.QueryRowContext(ctx, `
		SELECT focus_duration / 60, short_break_duration / 60
		FROM timer_settings WHERE user_id = $1`, userID,
	).Scan(&t.WorkMinutes, &t.RestMinutes); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var status, phase string
	var startedAt, completedAt, runningSince sql.NullTime
	var elapsedMs int64
	err := r.db.QueryRowContext(ctx, `
		SELECT status, phase, running_since, elapsed_ms, started_at, completed_at
		FROM timer_sessions WHERE timer_session_id = $1 AND user_id = $2`, sessionID, userID,
	).Scan(&status, &phase, &runningSince, &elapsedMs, &startedAt, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return t, nil
	}
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t.LastUpdated = startedAt.Time
	}
	if completedAt.Valid {
		t.LastUpdated = completedAt.Time
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM timer_cycles
		WHERE timer_session_id = $1 AND is_completed`, sessionID,
	).Scan(&t.CurrentCycle); err != nil {
		return nil, err
	}

	switch status {
	case "FOCUS", "SHORT_BREAK", "LONG_BREAK":
		t.Status = domain.StatusRunning
	default: // PAUSED, STOPPED, COMPLETED
		t.Status = domain.TimerStatus(status)
	}
	t.Phase = domain.TimerPhase(phase)
	if runningSince.Valid {
		t.RunningSince = runningSince.Time
	}
	t.Elapsed = time.Duration(elapsedMs) * time.Millisecond

	return t, nil
}

func (r *PostgresRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO timer_settings (user_id, focus_duration, short_break_duration)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET focus_duration = EXCLUDED.focus_duration,
		    short_break_duration = EXCLUDED.short_break_duration,
		    updated_at = CURRENT_TIMESTAMP`,
		t.UserID, t.WorkMinutes*60, t.RestMinutes*60,
	); err != nil {
		return err
	}

	dbStatus := string(t.Status)
	if t.Status == domain.StatusRunning {
		if t.Phase == domain.PhaseWork {
			dbStatus = "FOCUS"
		} else {
			dbStatus = "SHORT_BREAK"
		}
	}

	var completedAt any
	if t.Status == domain.StatusStopped {
		completedAt = t.LastUpdated
	}

	var runningSince any
	if t.Status == domain.StatusRunning {
		runningSince = t.RunningSince
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO timer_sessions (timer_session_id, user_id, status, phase, running_since, elapsed_ms, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (timer_session_id) DO UPDATE
		SET status = EXCLUDED.status,
		    phase = EXCLUDED.phase,
		    running_since = EXCLUDED.running_since,
		    elapsed_ms = EXCLUDED.elapsed_ms,
		    completed_at = EXCLUDED.completed_at`,
		t.SessionID, t.UserID, dbStatus, string(t.Phase), runningSince, t.Elapsed.Milliseconds(), t.LastUpdated, completedAt,
	); err != nil {
		return err
	}

	// A completed cycle is recorded once, when CurrentCycle advances past the
	// count of cycles already persisted for this timer session. The phase
	// that just finished is the opposite of the (already flipped) new phase.
	var recorded int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM timer_cycles WHERE timer_session_id = $1`, t.SessionID,
	).Scan(&recorded); err != nil {
		return err
	}

	for cycleNumber := recorded + 1; cycleNumber <= t.CurrentCycle; cycleNumber++ {
		phaseType, duration := "SHORT_BREAK", t.RestMinutes*60
		if t.Phase == domain.PhaseRest { // now resting => a FOCUS cycle just completed
			phaseType, duration = "FOCUS", t.WorkMinutes*60
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO timer_cycles (timer_session_id, cycle_number, phase_type, duration, is_completed, ended_at)
			VALUES ($1, $2, $3, $4, TRUE, $5)`,
			t.SessionID, cycleNumber, phaseType, duration, t.LastUpdated,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	h := &domain.TimerHistory{UserID: userID}

	var lastSessionAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(started_at)
		FROM timer_sessions WHERE user_id = $1`, userID,
	).Scan(&h.SessionsJoined, &lastSessionAt); err != nil {
		return nil, err
	}
	if lastSessionAt.Valid {
		h.LastActive = lastSessionAt.Time
	}

	var focusSeconds int
	var lastCycleAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FILTER (WHERE tc.is_completed),
		       COALESCE(SUM(tc.duration) FILTER (WHERE tc.is_completed AND tc.phase_type = 'FOCUS'), 0),
		       MAX(tc.ended_at)
		FROM timer_cycles tc
		JOIN timer_sessions ts ON ts.timer_session_id = tc.timer_session_id
		WHERE ts.user_id = $1`, userID,
	).Scan(&h.CyclesCompleted, &focusSeconds, &lastCycleAt); err != nil {
		return nil, err
	}
	h.TotalFocusMinutes = focusSeconds / 60

	if lastCycleAt.Valid && lastCycleAt.Time.After(h.LastActive) {
		h.LastActive = lastCycleAt.Time
	}

	daily, err := r.dailyFocusLast30Days(ctx, userID)
	if err != nil {
		return nil, err
	}
	h.DailyFocusMinutes = daily

	return h, nil
}

func (r *PostgresRepository) dailyFocusLast30Days(ctx context.Context, userID string) ([]domain.DailyFocus, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tc.ended_at::date AS day, SUM(tc.duration) / 60
		FROM timer_cycles tc
		JOIN timer_sessions ts ON ts.timer_session_id = tc.timer_session_id
		WHERE ts.user_id = $1
		  AND tc.is_completed
		  AND tc.phase_type = 'FOCUS'
		  AND tc.ended_at >= CURRENT_DATE - INTERVAL '29 days'
		GROUP BY day`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	minutesByDay := make(map[string]int)
	for rows.Next() {
		var day time.Time
		var minutes int
		if err := rows.Scan(&day, &minutes); err != nil {
			return nil, err
		}
		minutesByDay[day.Format("2006-01-02")] = minutes
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	daily := make([]domain.DailyFocus, 30)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := range 30 {
		date := today.AddDate(0, 0, -(29 - i)).Format("2006-01-02")
		daily[i] = domain.DailyFocus{Date: date, FocusMinutes: minutesByDay[date]}
	}
	return daily, nil
}

func (r *PostgresRepository) GetRoomTimers(ctx context.Context, sessionID string) ([]*domain.TimerState, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id::text FROM timers WHERE session_id = $1 AND status = 'OPEN'
		UNION
		SELECT user_id::text FROM timer_sessions WHERE timer_session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timers []*domain.TimerState
	var userIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, uid := range userIDs {
		t, err := r.GetTimer(ctx, sessionID, uid)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		if t != nil {
			timers = append(timers, t)
		}
	}
	return timers, nil
}

func (r *PostgresRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM processed_events WHERE event_id = $1`, eventID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresRepository) FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, eventType string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Idempotency Check: insert event_id. If duplicate, return nil immediately.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO processed_events (event_id, event_type)
		VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING`, eventID, eventType)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}

	now := time.Now().UTC()

	// 2. Discard in-progress cycles
	_, err = tx.ExecContext(ctx, `
		UPDATE cycles
		SET status = 'DISCARDED', ended_at = $1
		WHERE timer_id IN (
			SELECT timer_id FROM timers WHERE session_id = $2 AND user_id = $3
		) AND status IN ('RUNNING', 'PAUSED')`, now, sessionID, userID)
	if err != nil {
		return err
	}

	// 3. Mark timer as FINALIZED
	_, err = tx.ExecContext(ctx, `
		UPDATE timers
		SET status = 'FINALIZED', finalized_at = $1
		WHERE session_id = $2 AND user_id = $3 AND status = 'OPEN'`, now, sessionID, userID)
	if err != nil {
		return err
	}

	// 4. Update legacy timer_sessions table
	_, err = tx.ExecContext(ctx, `
		UPDATE timer_sessions
		SET status = 'STOPPED', completed_at = $1
		WHERE timer_session_id = $2 AND user_id = $3`, now, sessionID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) FinalizeSessionTimers(ctx context.Context, sessionID, eventID, eventType string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Idempotency Check
	res, err := tx.ExecContext(ctx, `
		INSERT INTO processed_events (event_id, event_type)
		VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING`, eventID, eventType)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}

	now := time.Now().UTC()

	// 2. Discard all in-progress cycles in this session
	_, err = tx.ExecContext(ctx, `
		UPDATE cycles
		SET status = 'DISCARDED', ended_at = $1
		WHERE timer_id IN (
			SELECT timer_id FROM timers WHERE session_id = $2
		) AND status IN ('RUNNING', 'PAUSED')`, now, sessionID)
	if err != nil {
		return err
	}

	// 3. Mark all timers in session as FINALIZED
	_, err = tx.ExecContext(ctx, `
		UPDATE timers
		SET status = 'FINALIZED', finalized_at = $1
		WHERE session_id = $2 AND status = 'OPEN'`, now, sessionID)
	if err != nil {
		return err
	}

	// 4. Update legacy timer_sessions table
	_, err = tx.ExecContext(ctx, `
		UPDATE timer_sessions
		SET status = 'STOPPED', completed_at = $1
		WHERE timer_session_id = $2`, now, sessionID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
