package repository

import (
	"context"
	"github.com/xmaximix/envilope-chako-server/internal/auth/domain"
	errs "github.com/xmaximix/envilope-chako-server/pkg/error"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, verified, role, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		u.ID, u.Email, u.PasswordHash, u.Verified, u.Role, u.CreatedAt,
	)
	return errs.Wrap("inserting user record", err)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	return &u, errs.Wrap("finding user by email", err)
}

func (r *UserRepo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES ($1,$2,$3)`,
		token, userID, expiresAt,
	)
	return errs.Wrap("storing refresh token", err)
}

func (r *UserRepo) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	var userID uuid.UUID
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, errs.Wrap("beginning transaction", err)
	}

	err = tx.GetContext(ctx, &userID, `SELECT user_id FROM refresh_tokens WHERE token = $1 AND expires_at > NOW()`, token)
	if err != nil {
		tx.Rollback()
		return uuid.Nil, errs.Wrap("validating refresh token", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, token)
	if err != nil {
		tx.Rollback()
		return uuid.Nil, errs.Wrap("deleting refresh token", err)
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, errs.Wrap("committing transaction", err)
	}
	return userID, nil
}
