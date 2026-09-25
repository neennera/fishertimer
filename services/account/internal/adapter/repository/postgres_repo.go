package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// PostgresRepository persists accounts in account_db.users
// (see database/schemas/001_create_users_table.sql).
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const userColumns = `user_id, email, display_name, avatar_url, role, created_at, updated_at`

func (r *PostgresRepository) CreateUser(ctx context.Context, u *domain.UserAccount) error {
	if u == nil || u.UserID == "" {
		return domain.ErrInvalid
	}
	const query = `
		INSERT INTO users (user_id, email, display_name, avatar_url, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		u.UserID, strings.ToLower(u.Email), u.DisplayName, u.AvatarURL, u.Role, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	return r.findOne(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, strings.ToLower(email))
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, userID string) (*domain.UserAccount, error) {
	return r.findOne(ctx, `SELECT `+userColumns+` FROM users WHERE user_id = $1`, userID)
}

func (r *PostgresRepository) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	u, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domain.Profile{
		UserID:      u.UserID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (r *PostgresRepository) UpdateProfile(ctx context.Context, p *domain.Profile) error {
	if p == nil || p.UserID == "" {
		return domain.ErrInvalid
	}
	const query = `UPDATE users SET display_name = $2, updated_at = $3 WHERE user_id = $1`
	res, err := r.db.ExecContext(ctx, query, p.UserID, p.DisplayName, p.UpdatedAt)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg any) (*domain.UserAccount, error) {
	var u domain.UserAccount
	var avatar sql.NullString

	err := r.db.QueryRowContext(ctx, query, arg).
		Scan(&u.UserID, &u.Email, &u.DisplayName, &avatar, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u.AvatarURL = avatar.String
	return &u, nil
}
