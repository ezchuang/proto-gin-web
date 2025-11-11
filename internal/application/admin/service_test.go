package admin

import (
	"context"
	"errors"
	"testing"

	authdomain "proto-gin-web/internal/admin/auth/domain"
)

type mockAdminRepo struct {
	adminByEmail    map[string]authdomain.StoredAdmin
	roleByName      map[string]authdomain.AdminRole
	createErr       error
	updateErr       error
	getErr          error
	roleErr         error
	createInput     authdomain.AdminCreateParams
	updateInput     authdomain.AdminProfileUpdateParams
	lastEmail       string
	lastUpdateEmail string
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{
		adminByEmail: map[string]authdomain.StoredAdmin{},
		roleByName:   map[string]authdomain.AdminRole{"admin": {ID: 1, Name: "admin"}},
	}
}

func (m *mockAdminRepo) GetByEmail(ctx context.Context, email string) (authdomain.StoredAdmin, error) {
	m.lastEmail = email
	if m.getErr != nil {
		return authdomain.StoredAdmin{}, m.getErr
	}
	if admin, ok := m.adminByEmail[email]; ok {
		return admin, nil
	}
	return authdomain.StoredAdmin{}, authdomain.ErrAdminNotFound
}

func (m *mockAdminRepo) Create(ctx context.Context, params authdomain.AdminCreateParams) (authdomain.StoredAdmin, error) {
	if m.createErr != nil {
		return authdomain.StoredAdmin{}, m.createErr
	}
	m.createInput = params
	admin := authdomain.StoredAdmin{
		Admin: authdomain.Admin{
			ID:          2,
			Email:       params.Email,
			DisplayName: params.DisplayName,
		},
		PasswordHash: params.PasswordHash,
	}
	m.adminByEmail[params.Email] = admin
	return admin, nil
}

func (m *mockAdminRepo) UpdateProfile(ctx context.Context, email string, params authdomain.AdminProfileUpdateParams) (authdomain.StoredAdmin, error) {
	if m.updateErr != nil {
		return authdomain.StoredAdmin{}, m.updateErr
	}
	m.lastUpdateEmail = email
	m.updateInput = params
	admin := m.adminByEmail[email]
	admin.DisplayName = params.DisplayName
	if params.PasswordHash != nil {
		admin.PasswordHash = *params.PasswordHash
	}
	m.adminByEmail[email] = admin
	return admin, nil
}

func (m *mockAdminRepo) FindRoleByName(ctx context.Context, name string) (authdomain.AdminRole, error) {
	if m.roleErr != nil {
		return authdomain.AdminRole{}, m.roleErr
	}
	if role, ok := m.roleByName[name]; ok {
		return role, nil
	}
	return authdomain.AdminRole{}, authdomain.ErrAdminRoleNotFound
}

func TestService_Login_Success(t *testing.T) {
	repo := newMockAdminRepo()
	hash, err := hashArgon2idPassword("pass1234")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repo.adminByEmail["user@example.com"] = authdomain.StoredAdmin{
		Admin: authdomain.Admin{
			ID:          1,
			Email:       "user@example.com",
			DisplayName: "User",
		},
		PasswordHash: hash,
	}
	svc := NewService(repo, Config{})

	admin, err := svc.Login(context.Background(), authdomain.AdminLoginInput{
		Email:    "user@example.com",
		Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if admin.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com got %s", admin.Email)
	}
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockAdminRepo()
	svc := NewService(repo, Config{})

	_, err := svc.Login(context.Background(), authdomain.AdminLoginInput{
		Email:    "missing@example.com",
		Password: "pass",
	})
	if !errors.Is(err, authdomain.ErrAdminInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestService_Login_LegacyFallback(t *testing.T) {
	repo := newMockAdminRepo()
	repo.getErr = authdomain.ErrAdminNotFound
	svc := NewService(repo, Config{
		LegacyUser:     "legacy@example.com",
		LegacyPassword: "legacy",
	})

	admin, err := svc.Login(context.Background(), authdomain.AdminLoginInput{
		Email:    "legacy@example.com",
		Password: "legacy",
	})
	if err != nil {
		t.Fatalf("expected legacy login to succeed: %v", err)
	}
	if admin.Email != "legacy@example.com" {
		t.Fatalf("unexpected legacy admin email %s", admin.Email)
	}
}

func TestService_Register_Validation(t *testing.T) {
	svc := NewService(newMockAdminRepo(), Config{})
	_, err := svc.Register(context.Background(), authdomain.AdminRegisterInput{
		Email:           "invalid",
		Password:        "12345678",
		ConfirmPassword: "12345678",
	})
	if !errors.Is(err, authdomain.ErrAdminInvalidEmail) {
		t.Fatalf("expected email error, got %v", err)
	}

	_, err = svc.Register(context.Background(), authdomain.AdminRegisterInput{
		Email:           "user@example.com",
		Password:        "short",
		ConfirmPassword: "short",
	})
	if !errors.Is(err, authdomain.ErrAdminPasswordTooShort) {
		t.Fatalf("expected short password error, got %v", err)
	}

	_, err = svc.Register(context.Background(), authdomain.AdminRegisterInput{
		Email:           "user@example.com",
		Password:        "longenough",
		ConfirmPassword: "different",
	})
	if !errors.Is(err, authdomain.ErrAdminPasswordMismatch) {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestService_Register_Success(t *testing.T) {
	repo := newMockAdminRepo()
	svc := NewService(repo, Config{})

	admin, err := svc.Register(context.Background(), authdomain.AdminRegisterInput{
		Email:           "user@example.com",
		Password:        "longenough",
		ConfirmPassword: "longenough",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if admin.Email != "user@example.com" {
		t.Fatalf("expected user@example.com got %s", admin.Email)
	}
	if repo.createInput.Email != "user@example.com" {
		t.Fatalf("expected repo create email to be set")
	}
}

func TestService_UpdateProfile_Validation(t *testing.T) {
	repo := newMockAdminRepo()
	repo.adminByEmail["user@example.com"] = authdomain.StoredAdmin{
		Admin: authdomain.Admin{Email: "user@example.com", DisplayName: "User"},
	}
	svc := NewService(repo, Config{})

	_, err := svc.UpdateProfile(context.Background(), "user@example.com", authdomain.AdminProfileInput{
		DisplayName: "",
	})
	if !errors.Is(err, authdomain.ErrAdminDisplayNameRequired) {
		t.Fatalf("expected display name required, got %v", err)
	}

	_, err = svc.UpdateProfile(context.Background(), "user@example.com", authdomain.AdminProfileInput{
		DisplayName:     "User",
		Password:        "short",
		ConfirmPassword: "short",
	})
	if !errors.Is(err, authdomain.ErrAdminPasswordTooShort) {
		t.Fatalf("expected short password error, got %v", err)
	}
}

func TestService_UpdateProfile_Success(t *testing.T) {
	repo := newMockAdminRepo()
	hash, _ := hashArgon2idPassword("pass1234")
	repo.adminByEmail["user@example.com"] = authdomain.StoredAdmin{
		Admin:        authdomain.Admin{Email: "user@example.com", DisplayName: "Old"},
		PasswordHash: hash,
	}
	svc := NewService(repo, Config{})

	admin, err := svc.UpdateProfile(context.Background(), "user@example.com", authdomain.AdminProfileInput{
		DisplayName:     "New",
		Password:        "newpassword",
		ConfirmPassword: "newpassword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if admin.DisplayName != "New" {
		t.Fatalf("expected updated display name, got %s", admin.DisplayName)
	}
	if repo.updateInput.DisplayName != "New" {
		t.Fatalf("repo update input not set")
	}
	if repo.updateInput.PasswordHash == nil {
		t.Fatalf("expected password hash to be set")
	}
}
