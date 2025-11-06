package admin

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"

	"proto-gin-web/internal/domain"
)

const (
	defaultPasswordMinLength = 8
	defaultAdminRoleName     = "admin"
)

// Config exposes optional knobs impacting admin behaviour.
type Config struct {
	AdminRoleName     string
	LegacyUser        string
	LegacyPassword    string
	PasswordMinLength int
}

// Service implements admin use cases.
type Service struct {
	repo domain.AdminRepository
	cfg  Config
}

var _ domain.AdminService = (*Service)(nil)

// NewService creates an admin service backed by a repository.
func NewService(repo domain.AdminRepository, cfg Config) *Service {
	if cfg.AdminRoleName == "" {
		cfg.AdminRoleName = defaultAdminRoleName
	}
	if cfg.PasswordMinLength <= 0 {
		cfg.PasswordMinLength = defaultPasswordMinLength
	}
	return &Service{repo: repo, cfg: cfg}
}

// Login authenticates an admin account.
func (s *Service) Login(ctx context.Context, input domain.AdminLoginInput) (domain.Admin, error) {
	email := domain.NormalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" {
		return domain.Admin{}, domain.ErrAdminInvalidCredentials
	}

	stored, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrAdminNotFound) && s.legacyMatches(email, password) {
			return s.legacyAdmin(email), nil
		}
		if errors.Is(err, domain.ErrAdminNotFound) {
			return domain.Admin{}, domain.ErrAdminInvalidCredentials
		}
		return domain.Admin{}, err
	}

	ok, verifyErr := verifyArgon2idHash(stored.PasswordHash, password)
	if verifyErr != nil {
		return domain.Admin{}, verifyErr
	}
	if !ok {
		return domain.Admin{}, domain.ErrAdminInvalidCredentials
	}
	return stored.Admin, nil
}

// Register provisions a new admin account.
func (s *Service) Register(ctx context.Context, input domain.AdminRegisterInput) (domain.Admin, error) {
	email := domain.NormalizeEmail(input.Email)
	if email == "" || !isLikelyEmail(email) {
		return domain.Admin{}, domain.ErrAdminInvalidEmail
	}
	password := strings.TrimSpace(input.Password)
	confirm := strings.TrimSpace(input.ConfirmPassword)
	if password == "" || confirm == "" {
		return domain.Admin{}, domain.ErrAdminPasswordTooShort
	}
	if len(password) < s.cfg.PasswordMinLength {
		return domain.Admin{}, domain.ErrAdminPasswordTooShort
	}
	if password != confirm {
		return domain.Admin{}, domain.ErrAdminPasswordMismatch
	}

	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return domain.Admin{}, domain.ErrAdminEmailExists
	} else if !errors.Is(err, domain.ErrAdminNotFound) {
		return domain.Admin{}, err
	}

	role, err := s.repo.FindRoleByName(ctx, s.cfg.AdminRoleName)
	if err != nil {
		return domain.Admin{}, err
	}

	hash, err := hashArgon2idPassword(password)
	if err != nil {
		return domain.Admin{}, err
	}

	display := deriveDisplayName(email)
	stored, err := s.repo.Create(ctx, domain.AdminCreateParams{
		Email:        email,
		DisplayName:  display,
		PasswordHash: hash,
		RoleID:       role.ID,
	})
	if err != nil {
		return domain.Admin{}, err
	}
	return stored.Admin, nil
}

// GetProfile fetches admin profile details.
func (s *Service) GetProfile(ctx context.Context, email string) (domain.Admin, error) {
	normalized := domain.NormalizeEmail(email)
	if normalized == "" {
		return domain.Admin{}, domain.ErrAdminNotFound
	}
	stored, err := s.repo.GetByEmail(ctx, normalized)
	if err != nil {
		return domain.Admin{}, err
	}
	return stored.Admin, nil
}

// UpdateProfile updates display name and optionally password.
func (s *Service) UpdateProfile(ctx context.Context, email string, input domain.AdminProfileInput) (domain.Admin, error) {
	normalized := domain.NormalizeEmail(email)
	if normalized == "" {
		return domain.Admin{}, domain.ErrAdminNotFound
	}
	display := strings.TrimSpace(input.DisplayName)
	if display == "" {
		return domain.Admin{}, domain.ErrAdminDisplayNameRequired
	}

	var passwordHash *string
	pass := strings.TrimSpace(input.Password)
	confirm := strings.TrimSpace(input.ConfirmPassword)
	if pass != "" || confirm != "" {
		if pass != confirm {
			return domain.Admin{}, domain.ErrAdminPasswordMismatch
		}
		if len(pass) < s.cfg.PasswordMinLength {
			return domain.Admin{}, domain.ErrAdminPasswordTooShort
		}
		hash, err := hashArgon2idPassword(pass)
		if err != nil {
			return domain.Admin{}, err
		}
		passwordHash = &hash
	}

	stored, err := s.repo.UpdateProfile(ctx, normalized, domain.AdminProfileUpdateParams{
		DisplayName:  display,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return domain.Admin{}, err
	}
	return stored.Admin, nil
}

func (s *Service) legacyMatches(email, password string) bool {
	if s.cfg.LegacyUser == "" || s.cfg.LegacyPassword == "" {
		return false
	}
	return strings.EqualFold(email, s.cfg.LegacyUser) && password == s.cfg.LegacyPassword
}

func (s *Service) legacyAdmin(email string) domain.Admin {
	display := strings.TrimSpace(s.cfg.LegacyUser)
	if display == "" {
		display = email
	}
	return domain.Admin{
		Email:       email,
		DisplayName: display,
	}
}

func hashArgon2idPassword(password string) (string, error) {
	const (
		time    = 2
		memory  = 64 * 1024
		threads = 1
		keyLen  = 32
		saltLen = 16
	)
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, time, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
	return encoded, nil
}

func verifyArgon2idHash(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid argon2 hash format")
	}
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported argon2 variant: %s", parts[1])
	}
	if !strings.HasPrefix(parts[2], "v=") {
		return false, fmt.Errorf("invalid argon2 version segment: %s", parts[2])
	}
	memory, iterations, parallelism, err := parseArgon2Params(parts[3])
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}
	computed := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	if len(computed) != len(expected) {
		return false, fmt.Errorf("argon2 hash length mismatch")
	}
	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}

func parseArgon2Params(segment string) (memory uint32, iterations uint32, parallelism uint8, err error) {
	fields := strings.Split(segment, ",")
	for _, field := range fields {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) != 2 {
			return 0, 0, 0, fmt.Errorf("invalid argon2 param: %s", field)
		}
		value, parseErr := strconv.ParseUint(kv[1], 10, 32)
		if parseErr != nil {
			return 0, 0, 0, fmt.Errorf("parse argon2 param %s: %w", kv[0], parseErr)
		}
		switch kv[0] {
		case "m":
			memory = uint32(value)
		case "t":
			iterations = uint32(value)
		case "p":
			parallelism = uint8(value)
		}
	}
	if memory == 0 || iterations == 0 || parallelism == 0 {
		return 0, 0, 0, fmt.Errorf("argon2 parameters incomplete")
	}
	return memory, iterations, parallelism, nil
}

func isLikelyEmail(input string) bool {
	if input == "" || strings.Count(input, "@") != 1 {
		return false
	}
	local, domainPart, ok := strings.Cut(input, "@")
	if !ok || local == "" || domainPart == "" {
		return false
	}
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") || !strings.Contains(domainPart, ".") {
		return false
	}
	return true
}

func deriveDisplayName(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	local := email[:at]
	parts := strings.FieldsFunc(local, func(r rune) bool {
		return r == '.' || r == '_' || r == '-' || r == '+'
	})
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	if len(parts) == 0 {
		return email
	}
	return strings.Join(parts, " ")
}
