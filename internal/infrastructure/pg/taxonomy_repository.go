package pg

import (
	"context"

	"proto-gin-web/internal/domain"
)

// TaxonomyRepository implements domain.TaxonomyRepository backed by pgx queries.
type TaxonomyRepository struct {
	queries *Queries
}

// NewTaxonomyRepository creates a taxonomy repository.
func NewTaxonomyRepository(queries *Queries) *TaxonomyRepository {
	return &TaxonomyRepository{queries: queries}
}

var _ domain.TaxonomyRepository = (*TaxonomyRepository)(nil)

func (r *TaxonomyRepository) CreateCategory(ctx context.Context, input domain.CreateCategoryInput) (domain.Category, error) {
	row, err := r.queries.CreateCategory(ctx, CreateCategoryParams{Name: input.Name, Slug: input.Slug})
	if err != nil {
		return domain.Category{}, err
	}
	return domain.Category{
		ID:   row.ID,
		Name: row.Name,
		Slug: row.Slug,
	}, nil
}

func (r *TaxonomyRepository) DeleteCategory(ctx context.Context, slug string) error {
	return r.queries.DeleteCategoryBySlug(ctx, slug)
}

func (r *TaxonomyRepository) CreateTag(ctx context.Context, input domain.CreateTagInput) (domain.Tag, error) {
	row, err := r.queries.CreateTag(ctx, CreateTagParams{Name: input.Name, Slug: input.Slug})
	if err != nil {
		return domain.Tag{}, err
	}
	return domain.Tag{
		ID:   row.ID,
		Name: row.Name,
		Slug: row.Slug,
	}, nil
}

func (r *TaxonomyRepository) DeleteTag(ctx context.Context, slug string) error {
	return r.queries.DeleteTagBySlug(ctx, slug)
}
