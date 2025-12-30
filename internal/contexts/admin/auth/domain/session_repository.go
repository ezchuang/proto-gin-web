package authdomain

import (
	"context"
	"time"
)

// SessionStore persists admin sessions.
type SessionStore interface {
	Save(ctx context.Context, session AdminSession, ttl time.Duration) error
	Get(ctx context.Context, id string) (AdminSession, error)
	Touch(ctx context.Context, session AdminSession, ttl time.Duration) error
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, userID int64) error
}

// RememberTokenRepository persists remember tokens.
type RememberTokenRepository interface {
	Insert(ctx context.Context, token RememberToken) error
	GetBySelector(ctx context.Context, selector string) (RememberToken, error)
	Update(ctx context.Context, token RememberToken) error
	Delete(ctx context.Context, selector string) error
	DeleteByUser(ctx context.Context, userID int64) error
}
