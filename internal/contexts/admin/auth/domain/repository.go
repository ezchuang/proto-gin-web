package authdomain

import "context"

// AdminRepository abstracts persistence concerns for admin accounts.
type AdminRepository interface {
	GetByEmail(ctx context.Context, email string) (StoredAdmin, error)
	GetByID(ctx context.Context, id int64) (StoredAdmin, error)
	Create(ctx context.Context, params AdminCreateParams) (StoredAdmin, error)
	UpdateProfile(ctx context.Context, email string, params AdminProfileUpdateParams) (StoredAdmin, error)
	FindRoleByName(ctx context.Context, role string) (AdminRole, error)
}
