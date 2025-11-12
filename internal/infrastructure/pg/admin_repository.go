package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	authdomain "proto-gin-web/internal/admin/auth/domain"
)

// AdminAccountRepository implements authdomain.AdminRepository using pgx Queries.
type AdminAccountRepository struct {
	queries *Queries
}

// NewAdminAccountRepository constructs a repository using the shared Queries helper.
func NewAdminAccountRepository(queries *Queries) *AdminAccountRepository {
	return &AdminAccountRepository{queries: queries}
}

var _ authdomain.AdminRepository = (*AdminAccountRepository)(nil)

func (r *AdminAccountRepository) GetByEmail(ctx context.Context, email string) (authdomain.StoredAdmin, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.StoredAdmin{}, authdomain.ErrAdminNotFound
		}
		return authdomain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) GetByID(ctx context.Context, id int64) (authdomain.StoredAdmin, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.StoredAdmin{}, authdomain.ErrAdminNotFound
		}
		return authdomain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) Create(ctx context.Context, params authdomain.AdminCreateParams) (authdomain.StoredAdmin, error) {
	roleID := params.RoleID
	user, err := r.queries.CreateUser(ctx, CreateUserParams{
		Email:        params.Email,
		DisplayName:  params.DisplayName,
		PasswordHash: params.PasswordHash,
		RoleID:       &roleID,
	})
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return authdomain.StoredAdmin{}, authdomain.ErrAdminEmailExists
		}
		return authdomain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) UpdateProfile(ctx context.Context, email string, params authdomain.AdminProfileUpdateParams) (authdomain.StoredAdmin, error) {
	user, err := r.queries.UpdateUserProfile(ctx, UpdateUserProfileParams{
		Email:        email,
		DisplayName:  params.DisplayName,
		PasswordHash: params.PasswordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.StoredAdmin{}, authdomain.ErrAdminNotFound
		}
		return authdomain.StoredAdmin{}, err
	}
	return mapStoredAdmin(user), nil
}

func (r *AdminAccountRepository) FindRoleByName(ctx context.Context, name string) (authdomain.AdminRole, error) {
	role, err := r.queries.GetRoleByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.AdminRole{}, authdomain.ErrAdminRoleNotFound
		}
		return authdomain.AdminRole{}, err
	}
	return authdomain.AdminRole{ID: role.ID, Name: role.Name}, nil
}

func mapStoredAdmin(u User) authdomain.StoredAdmin {
	var roleID *int64
	if u.RoleID.Valid {
		roleID = ptr(u.RoleID.Int64)
	}
	return authdomain.StoredAdmin{
		Admin: authdomain.Admin{
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
