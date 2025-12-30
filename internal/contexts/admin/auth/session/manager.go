package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"time"

	authdomain "proto-gin-web/internal/contexts/admin/auth/domain"
)

// Config holds lifetimes for sessions and remember-me cookies.
type Config struct {
	SessionTTL         time.Duration
	SessionAbsoluteTTL time.Duration
	RememberTTL        time.Duration
}

// Manager coordinates Redis-backed sessions and remember tokens.
type Manager struct {
	store  authdomain.SessionStore
	tokens authdomain.RememberTokenRepository
	cfg    Config
}

// RememberTokenSecret contains the clear-text split token pieces returned to the caller.
type RememberTokenSecret struct {
	Selector  string
	Validator string
	ExpiresAt time.Time
}

// NewManager constructs a Manager with sane defaults.
func NewManager(store authdomain.SessionStore, tokens authdomain.RememberTokenRepository, cfg Config) *Manager {
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = 30 * time.Minute
	}
	if cfg.SessionAbsoluteTTL <= 0 {
		cfg.SessionAbsoluteTTL = 12 * time.Hour
	}
	if cfg.RememberTTL <= 0 {
		cfg.RememberTTL = 30 * 24 * time.Hour
	}
	return &Manager{
		store:  store,
		tokens: tokens,
		cfg:    cfg,
	}
}

// IssueSession creates a new short-lived admin session.
func (m *Manager) IssueSession(ctx context.Context, admin authdomain.Admin) (authdomain.AdminSession, error) {
	now := time.Now().UTC()
	id, err := randomHex(32)
	if err != nil {
		return authdomain.AdminSession{}, err
	}
	session := authdomain.AdminSession{
		ID:             id,
		Profile:        admin,
		IssuedAt:       now,
		AbsoluteExpiry: now.Add(m.cfg.SessionAbsoluteTTL),
	}
	ttl := m.sessionTTL(session, now)
	if ttl <= 0 {
		return authdomain.AdminSession{}, authdomain.ErrAdminSessionExpired
	}
	if err := m.store.Save(ctx, session, ttl); err != nil {
		return authdomain.AdminSession{}, err
	}
	return session, nil
}

// ValidateSession loads a session and refreshes its TTL if still valid.
func (m *Manager) ValidateSession(ctx context.Context, id string) (authdomain.AdminSession, error) {
	session, err := m.store.Get(ctx, id)
	if err != nil {
		return authdomain.AdminSession{}, err
	}
	now := time.Now().UTC()
	if now.After(session.AbsoluteExpiry) {
		_ = m.store.Delete(ctx, id)
		return authdomain.AdminSession{}, authdomain.ErrAdminSessionExpired
	}
	ttl := m.sessionTTL(session, now)
	if err := m.store.Touch(ctx, session, ttl); err != nil {
		return authdomain.AdminSession{}, err
	}
	return session, nil
}

// RefreshSessionProfile updates the stored profile snapshot after profile edits.
func (m *Manager) RefreshSessionProfile(ctx context.Context, sessionID string, admin authdomain.Admin) error {
	session, err := m.store.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	session.Profile = admin
	now := time.Now().UTC()
	ttl := m.sessionTTL(session, now)
	return m.store.Save(ctx, session, ttl)
}

// DestroySession removes a specific session.
func (m *Manager) DestroySession(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

// DestroyAllSessions removes all sessions for the given user.
func (m *Manager) DestroyAllSessions(ctx context.Context, userID int64) error {
	return m.store.DeleteByUser(ctx, userID)
}

// CreateRememberToken issues a new remember-me token.
func (m *Manager) CreateRememberToken(ctx context.Context, userID int64, deviceInfo string) (RememberTokenSecret, error) {
	selector, err := randomBase64(16)
	if err != nil {
		return RememberTokenSecret{}, err
	}
	validator, err := randomBase64(32)
	if err != nil {
		return RememberTokenSecret{}, err
	}
	now := time.Now().UTC()
	token := authdomain.RememberToken{
		Selector:      selector,
		ValidatorHash: hashValidator(validator),
		UserID:        userID,
		DeviceInfo:    deviceInfo,
		ExpiresAt:     now.Add(m.cfg.RememberTTL),
		LastUsedAt:    now,
		Revoked:       false,
	}
	if err := m.tokens.Insert(ctx, token); err != nil {
		return RememberTokenSecret{}, err
	}
	return RememberTokenSecret{
		Selector:  selector,
		Validator: validator,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

// ValidateRememberToken verifies the split token, rotates the validator, and returns the stored row.
func (m *Manager) ValidateRememberToken(ctx context.Context, selector, validator, deviceInfo string) (authdomain.RememberToken, RememberTokenSecret, error) {
	token, err := m.tokens.GetBySelector(ctx, selector)
	if err != nil {
		return authdomain.RememberToken{}, RememberTokenSecret{}, err
	}
	if token.Revoked || time.Now().UTC().After(token.ExpiresAt) {
		_ = m.tokens.Delete(ctx, selector)
		return authdomain.RememberToken{}, RememberTokenSecret{}, authdomain.ErrAdminRememberTokenNotFound
	}
	hashed := hashValidator(validator)
	if !constantTimeEquals(hashed, token.ValidatorHash) {
		_ = m.tokens.Delete(ctx, selector)
		return authdomain.RememberToken{}, RememberTokenSecret{}, authdomain.ErrAdminRememberTokenInvalid
	}
	newValidator, err := randomBase64(32)
	if err != nil {
		return authdomain.RememberToken{}, RememberTokenSecret{}, err
	}
	token.ValidatorHash = hashValidator(newValidator)
	token.LastUsedAt = time.Now().UTC()
	token.ExpiresAt = token.LastUsedAt.Add(m.cfg.RememberTTL)
	token.DeviceInfo = deviceInfo
	token.Revoked = false
	if err := m.tokens.Update(ctx, token); err != nil {
		return authdomain.RememberToken{}, RememberTokenSecret{}, err
	}
	return token, RememberTokenSecret{
		Selector:  token.Selector,
		Validator: newValidator,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

// DeleteRememberToken removes a selector entry.
func (m *Manager) DeleteRememberToken(ctx context.Context, selector string) error {
	return m.tokens.Delete(ctx, selector)
}

// DeleteRememberTokensByUser removes all remember tokens for a user.
func (m *Manager) DeleteRememberTokensByUser(ctx context.Context, userID int64) error {
	return m.tokens.DeleteByUser(ctx, userID)
}

func (m *Manager) sessionTTL(session authdomain.AdminSession, now time.Time) time.Duration {
	remaining := session.AbsoluteExpiry.Sub(now)
	if remaining <= 0 {
		return 0
	}
	if remaining < m.cfg.SessionTTL {
		return remaining
	}
	return m.cfg.SessionTTL
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randomBase64(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashValidator(validator string) string {
	sum := sha256.Sum256([]byte(validator))
	return hex.EncodeToString(sum[:])
}

func constantTimeEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

