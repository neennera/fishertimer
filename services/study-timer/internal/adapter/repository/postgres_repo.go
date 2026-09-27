package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// PostgresRepository persists timer state in timer_db (timer_settings,
// timer_sessions, timer_cycles) - see
// database/schemas/001_create_timer_tables.sql.
//
// The schema's timer_sessions.status enum (FOCUS/SHORT_BREAK/LONG_BREAK/
// PAUSED/COMPLETED/STOPPED) captures the *current running phase* while a
// timer is active, collapsing to PAUSED/STOPPED/COMPLETED otherwise. Reads
// reconstruct domain.TimerState.Phase from that column when the timer is
// running, or from the most recently completed cycle's phase_type (its
// opposite - CompleteCycle always flips phase) when it is not.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t := &domain.TimerState{
		SessionID:   sessionID,
		UserID:      userID,
		Status:      "STOPPED",
		Phase:       domain.PhaseWork,
		WorkMinutes: 25,
		RestMinutes: 5,
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT focus_duration / 60, short_break_duration / 60
		FROM timer_settings WHERE user_id = $1`, userID,
	).Scan(&t.WorkMinutes, &t.RestMinutes); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var status string
	var startedAt, completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT status, started_at, completed_at
		FROM timer_sessions WHERE timer_session_id = $1 AND user_id = $2`, sessionID, userID,
	).Scan(&status, &startedAt, &completedAt)
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
	case "FOCUS":
		t.Status, t.Phase = "RUNNING", domain.PhaseWork
	case "SHORT_BREAK", "LONG_BREAK":
		t.Status, t.Phase = "RUNNING", domain.PhaseRest
	default: // PAUSED, STOPPED, COMPLETED
		t.Status = status
		t.Phase = r.lastCompletedPhaseOpposite(ctx, sessionID)
	}

	return t, nil
}

// lastCompletedPhaseOpposite looks at the most recently completed cycle to
// figure out which phase is live once a timer is no longer actively running
// (paused, stopped or completed) - CompleteCycle always flips to the other
// phase, so the live phase is the opposite of the last one recorded.
func (r *PostgresRepository) lastCompletedPhaseOpposite(ctx context.Context, sessionID string) domain.TimerPhase {
	var phaseType string
	err := r.db.QueryRowContext(ctx, `
		SELECT phase_type FROM timer_cycles
		WHERE timer_session_id = $1 AND is_completed
		ORDER BY ended_at DESC LIMIT 1`, sessionID,
	).Scan(&phaseType)
	if err != nil {
		return domain.PhaseWork
	}
	if phaseType == "FOCUS" {
		return domain.PhaseRest
	}
	return domain.PhaseWork
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

	dbStatus := t.Status
	if t.Status == "RUNNING" {
		if t.Phase == domain.PhaseWork {
			dbStatus = "FOCUS"
		} else {
			dbStatus = "SHORT_BREAK"
		}
	}

	var completedAt any
	if t.Status == "STOPPED" || t.Status == "COMPLETED" {
		completedAt = t.LastUpdated
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO timer_sessions (timer_session_id, user_id, status, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (timer_session_id) DO UPDATE
		SET status = EXCLUDED.status,
		    completed_at = EXCLUDED.completed_at`,
		t.SessionID, t.UserID, dbStatus, t.LastUpdated, completedAt,
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

	return h, nil
}
