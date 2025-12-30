package pg

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "proto-gin-web/internal/contexts/admin/auth/domain"
)

// RememberTokenRepository persists split-token metadata in Postgres.
type RememberTokenRepository struct {
	pool *pgxpool.Pool
}

// NewRememberTokenRepository constructs the repository.
func NewRememberTokenRepository(pool *pgxpool.Pool) *RememberTokenRepository {
	return &RememberTokenRepository{pool: pool}
}

var _ authdomain.RememberTokenRepository = (*RememberTokenRepository)(nil)

// Insert saves a new remember token row.
func (r *RememberTokenRepository) Insert(ctx context.Context, token authdomain.RememberToken) error {
	const stmt = `
		INSERT INTO admin_remember_tokens (selector, validator_hash, user_id, device_info, expires_at, last_used_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, stmt,
		token.Selector,
		token.ValidatorHash,
		token.UserID,
		nullIfEmpty(token.DeviceInfo),
		token.ExpiresAt,
		token.LastUsedAt,
		token.Revoked,
	)
	return err
}

// GetBySelector fetches a remember token by selector.
func (r *RememberTokenRepository) GetBySelector(ctx context.Context, selector string) (authdomain.RememberToken, error) {
	const stmt = `
		SELECT selector, validator_hash, user_id, device_info, expires_at, last_used_at, revoked
		FROM admin_remember_tokens
		WHERE selector = $1
	`
	var (
		token  authdomain.RememberToken
		device sql.NullString
	)
	err := r.pool.QueryRow(ctx, stmt, selector).Scan(
		&token.Selector,
		&token.ValidatorHash,
		&token.UserID,
		&device,
		&token.ExpiresAt,
		&token.LastUsedAt,
		&token.Revoked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.RememberToken{}, authdomain.ErrAdminRememberTokenNotFound
		}
		return authdomain.RememberToken{}, err
	}
	if device.Valid {
		token.DeviceInfo = device.String
	}
	return token, nil
}

// Update rotates validator metadata for an existing remember token.
func (r *RememberTokenRepository) Update(ctx context.Context, token authdomain.RememberToken) error {
	const stmt = `
		UPDATE admin_remember_tokens
		SET validator_hash = $2,
		    device_info = $3,
		    expires_at = $4,
		    last_used_at = $5,
		    revoked = $6
		WHERE selector = $1
	`
	ct, err := r.pool.Exec(ctx, stmt,
		token.Selector,
		token.ValidatorHash,
		nullIfEmpty(token.DeviceInfo),
		token.ExpiresAt,
		token.LastUsedAt,
		token.Revoked,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return authdomain.ErrAdminRememberTokenNotFound
	}
	return nil
}

// Delete removes a stored token by selector.
func (r *RememberTokenRepository) Delete(ctx context.Context, selector string) error {
	const stmt = `DELETE FROM admin_remember_tokens WHERE selector = $1`
	_, err := r.pool.Exec(ctx, stmt, selector)
	return err
}

// DeleteByUser removes all remember tokens for the provided user id.
func (r *RememberTokenRepository) DeleteByUser(ctx context.Context, userID int64) error {
	const stmt = `DELETE FROM admin_remember_tokens WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, stmt, userID)
	return err
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

