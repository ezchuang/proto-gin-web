package authdomain

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	// ErrAdminNotFound indicates the requested administrator account was not found.
	ErrAdminNotFound = errors.New("admin: account not found")
	// ErrAdminEmailExists signals an attempt to create an account with an existing email.
	ErrAdminEmailExists = errors.New("admin: email already exists")
	// ErrAdminInvalidCredentials occurs when login credentials are invalid.
	ErrAdminInvalidCredentials = errors.New("admin: invalid credentials")
	// ErrAdminInvalidEmail indicates the provided email address fails validation.
	ErrAdminInvalidEmail = errors.New("admin: invalid email address")
	// ErrAdminPasswordTooShort indicates the supplied password is shorter than the minimum requirement.
	ErrAdminPasswordTooShort = errors.New("admin: password must be at least 8 characters")
	// ErrAdminPasswordMismatch indicates password confirmation did not match.
	ErrAdminPasswordMismatch = errors.New("admin: passwords do not match")
	// ErrAdminDisplayNameRequired indicates display name input was empty.
	ErrAdminDisplayNameRequired = errors.New("admin: display name is required")
	// ErrAdminRoleNotFound indicates the role lookup failed.
	ErrAdminRoleNotFound = errors.New("admin: role not found")
	// ErrAdminSessionNotFound indicates the session was missing or expired.
	ErrAdminSessionNotFound = errors.New("admin: session not found")
	// ErrAdminSessionExpired indicates the session exceeded its absolute lifetime.
	ErrAdminSessionExpired = errors.New("admin: session expired")
	// ErrAdminRememberTokenNotFound indicates the remember-me selector was not found.
	ErrAdminRememberTokenNotFound = errors.New("admin: remember token not found")
	// ErrAdminRememberTokenInvalid indicates the validator failed constant-time comparison.
	ErrAdminRememberTokenInvalid = errors.New("admin: remember token invalid")
)

// Admin represents public administrator information.
type Admin struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	RoleID      *int64    `json:"role_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StoredAdmin contains persisted admin data including sensitive fields.
type StoredAdmin struct {
	Admin
	PasswordHash string
}

// AdminRole provides role metadata used when assigning admin accounts.
type AdminRole struct {
	ID   int64
	Name string
}

// AdminCreateParams contains data required to create an admin account.
type AdminCreateParams struct {
	Email        string
	DisplayName  string
	PasswordHash string
	RoleID       int64
}

// AdminProfileUpdateParams holds profile update values.
type AdminProfileUpdateParams struct {
	DisplayName  string
	PasswordHash *string
}

// AdminLoginInput represents login credentials.
type AdminLoginInput struct {
	Email    string
	Password string
}

// AdminRegisterInput captures information for registering a new admin.
type AdminRegisterInput struct {
	Email           string
	Password        string
	ConfirmPassword string
}

// AdminProfileInput contains profile changes submitted by an admin.
type AdminProfileInput struct {
	DisplayName     string
	Password        string
	ConfirmPassword string
}

// AdminRepository abstracts persistence concerns for admin accounts.
type AdminRepository interface {
	GetByEmail(ctx context.Context, email string) (StoredAdmin, error)
	GetByID(ctx context.Context, id int64) (StoredAdmin, error)
	Create(ctx context.Context, params AdminCreateParams) (StoredAdmin, error)
	UpdateProfile(ctx context.Context, email string, params AdminProfileUpdateParams) (StoredAdmin, error)
	FindRoleByName(ctx context.Context, role string) (AdminRole, error)
}

// NormalizeEmail lowercases and trims an email.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// AdminSession captures authenticated session data stored in Redis.
type AdminSession struct {
	ID             string
	Profile        Admin
	IssuedAt       time.Time
	AbsoluteExpiry time.Time
}

// SessionStore persists admin sessions.
type SessionStore interface {
	Save(ctx context.Context, session AdminSession, ttl time.Duration) error
	Get(ctx context.Context, id string) (AdminSession, error)
	Touch(ctx context.Context, session AdminSession, ttl time.Duration) error
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, userID int64) error
}

// RememberToken stores opaque split-token metadata.
type RememberToken struct {
	Selector      string
	ValidatorHash string
	UserID        int64
	DeviceInfo    string
	ExpiresAt     time.Time
	LastUsedAt    time.Time
	Revoked       bool
}

// RememberTokenRepository persists remember tokens.
type RememberTokenRepository interface {
	Insert(ctx context.Context, token RememberToken) error
	GetBySelector(ctx context.Context, selector string) (RememberToken, error)
	Update(ctx context.Context, token RememberToken) error
	Delete(ctx context.Context, selector string) error
	DeleteByUser(ctx context.Context, userID int64) error
}
