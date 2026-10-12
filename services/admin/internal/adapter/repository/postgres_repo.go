package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

// PostgresRepository persists moderation actions in admin_db.admin_logs
// (see database/schemas/001_create_admin_logs_table.sql).
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Record inserts the log row and runs apply inside one transaction, so the
// row commits only when the Study Session command succeeded, and a command is
// never sent while admin_db is unreachable (UC-10 E-4).
func (r *PostgresRepository) Record(ctx context.Context, entry *domain.AdminLog, apply func(ctx context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("admin: begin audit log: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO admin_logs (admin_id, action, target_id, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING log_id, created_at`,
		entry.AdminID, string(entry.Action), entry.TargetID, entry.Reason,
	).Scan(&entry.LogID, &entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("admin: insert audit log: %w", err)
	}

	if err := apply(ctx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("admin: commit audit log: %w", err)
	}
	return nil
}

func (r *PostgresRepository) List(ctx context.Context, action domain.AdminAction, limit int) ([]domain.AdminLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT log_id, admin_id, action, target_id, reason, created_at
		FROM admin_logs
		WHERE $1 = '' OR action = $1
		ORDER BY created_at DESC
		LIMIT $2`, string(action), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []domain.AdminLog{}
	for rows.Next() {
		var l domain.AdminLog
		var reason sql.NullString
		if err := rows.Scan(&l.LogID, &l.AdminID, &l.Action, &l.TargetID, &reason, &l.CreatedAt); err != nil {
			return nil, err
		}
		if reason.Valid {
			l.Reason = &reason.String
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
