package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"proto-gin-web/internal/domain"
)

// AdminAccountRepository implements domain.AdminRepository backed by pgx queries.
type AdminAccountRepository struct {
	queries *Queries
}

// NewAdminAccountRepository constructs a repository using the shared Queries helper.
func NewAdminAccountRepository(queries *Queries) *AdminAccountRepository {
	return &AdminAccountRepository{queries: queries}
}

var _ domain.AdminRepository = (*AdminAccountRepository)(nil)

func (r *AdminAccountRepository) GetByEmail(ctx context.Context, email string) (domain.StoredAdmin, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StoredAdmin{}, domain.ErrAdminNotFound
		}
		return domain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) Create(ctx context.Context, params domain.AdminCreateParams) (domain.StoredAdmin, error) {
	roleID := params.RoleID
	user, err := r.queries.CreateUser(ctx, CreateUserParams{
		Email:        params.Email,
		DisplayName:  params.DisplayName,
		PasswordHash: params.PasswordHash,
		RoleID:       &roleID,
	})
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return domain.StoredAdmin{}, domain.ErrAdminEmailExists
		}
		return domain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) UpdateProfile(ctx context.Context, email string, params domain.AdminProfileUpdateParams) (domain.StoredAdmin, error) {
	user, err := r.queries.UpdateUserProfile(ctx, UpdateUserProfileParams{
		Email:        email,
		DisplayName:  params.DisplayName,
		PasswordHash: params.PasswordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StoredAdmin{}, domain.ErrAdminNotFound
		}
		return domain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) FindRoleByName(ctx context.Context, name string) (domain.AdminRole, error) {
	role, err := r.queries.GetRoleByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AdminRole{}, domain.ErrAdminRoleNotFound
		}
		return domain.AdminRole{}, err
	}
	return domain.AdminRole{ID: role.ID, Name: role.Name}, nil
}

func mapStoredAdmin(u User) domain.StoredAdmin {
	var roleID *int64
	if u.RoleID.Valid {
		roleID = ptr(u.RoleID.Int64)
	}
	return domain.StoredAdmin{
		Admin: domain.Admin{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			RoleID:      roleID,
			CreatedAt:   u.CreatedAt,
		},
		PasswordHash: u.PasswordHash,
	}
}

func ptr[T any](v T) *T {
	return &v
}
