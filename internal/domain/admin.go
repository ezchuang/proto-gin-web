package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	// ErrAdminNotFound indicates the requested administrator account does not exist.
	ErrAdminNotFound = errors.New("admin: account not found")
	// ErrAdminEmailExists signals that an account already uses the provided email.
	ErrAdminEmailExists = errors.New("admin: email already exists")
	// ErrAdminInvalidCredentials is returned when login fails due to bad email/password combination.
	ErrAdminInvalidCredentials = errors.New("admin: invalid credentials")
	// ErrAdminInvalidEmail represents an invalid email format.
	ErrAdminInvalidEmail = errors.New("admin: invalid email address")
	// ErrAdminPasswordTooShort indicates the password does not meet length requirements.
	ErrAdminPasswordTooShort = errors.New("admin: password must be at least 8 characters")
	// ErrAdminPasswordMismatch indicates the confirmation password did not match.
	ErrAdminPasswordMismatch = errors.New("admin: passwords do not match")
	// ErrAdminDisplayNameRequired indicates display name input was empty.
	ErrAdminDisplayNameRequired = errors.New("admin: display name is required")
	// ErrAdminRoleNotFound indicates the target role record was not found.
	ErrAdminRoleNotFound = errors.New("admin: role not found")
)

// Admin models the public information for an administrator account.
type Admin struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	RoleID      *int64    `json:"role_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StoredAdmin includes sensitive fields retrieved from persistence.
type StoredAdmin struct {
	Admin
	PasswordHash string
}

// AdminRole models a user role row relevant to admin provisioning.
type AdminRole struct {
	ID   int64
	Name string
}

// AdminCreateParams captures data needed to persist a new admin account.
type AdminCreateParams struct {
	Email        string
	DisplayName  string
	PasswordHash string
	RoleID       int64
}

// AdminProfileUpdateParams describes updates to an existing account.
type AdminProfileUpdateParams struct {
	DisplayName  string
	PasswordHash *string
}

// AdminLoginInput describes credentials used for authentication.
type AdminLoginInput struct {
	Email    string
	Password string
}

// AdminRegisterInput provides details for registering a new administrator account.
type AdminRegisterInput struct {
	Email           string
	Password        string
	ConfirmPassword string
}

// AdminProfileInput captures profile edits submitted by the account owner.
type AdminProfileInput struct {
	DisplayName     string
	Password        string
	ConfirmPassword string
}

// AdminRepository abstracts persistence for administrator accounts.
type AdminRepository interface {
	GetByEmail(ctx context.Context, email string) (StoredAdmin, error)
	Create(ctx context.Context, params AdminCreateParams) (StoredAdmin, error)
	UpdateProfile(ctx context.Context, email string, params AdminProfileUpdateParams) (StoredAdmin, error)
	FindRoleByName(ctx context.Context, role string) (AdminRole, error)
}

// AdminService defines use-cases around administrator accounts.
type AdminService interface {
	Login(ctx context.Context, input AdminLoginInput) (Admin, error)
	Register(ctx context.Context, input AdminRegisterInput) (Admin, error)
	GetProfile(ctx context.Context, email string) (Admin, error)
	UpdateProfile(ctx context.Context, email string, input AdminProfileInput) (Admin, error)
}

// NormalizeEmail lowercases and trims an input email.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
