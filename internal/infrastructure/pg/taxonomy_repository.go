package pg

import (
	"context"

	taxdomain "proto-gin-web/internal/blog/taxonomy/domain"
)

// TaxonomyRepository implements taxdomain.TaxonomyRepository backed by pgx queries.
type TaxonomyRepository struct {
	queries *Queries
}

// NewTaxonomyRepository creates a taxonomy repository.
func NewTaxonomyRepository(queries *Queries) *TaxonomyRepository {
	return &TaxonomyRepository{queries: queries}
}

var _ taxdomain.TaxonomyRepository = (*TaxonomyRepository)(nil)

func (r *TaxonomyRepository) CreateCategory(ctx context.Context, input taxdomain.CreateCategoryInput) (taxdomain.Category, error) {
	row, err := r.queries.CreateCategory(ctx, CreateCategoryParams{Name: input.Name, Slug: input.Slug})
	if err != nil {
		return taxdomain.Category{}, err
	}
	return taxdomain.Category{
		ID:   row.ID,
		Name: row.Name,
		Slug: row.Slug,
	}, nil
}

func (r *TaxonomyRepository) DeleteCategory(ctx context.Context, slug string) error {
	return r.queries.DeleteCategoryBySlug(ctx, slug)
}

func (r *TaxonomyRepository) CreateTag(ctx context.Context, input taxdomain.CreateTagInput) (taxdomain.Tag, error) {
	row, err := r.queries.CreateTag(ctx, CreateTagParams{Name: input.Name, Slug: input.Slug})
	if err != nil {
		return taxdomain.Tag{}, err
	}
	return taxdomain.Tag{
		ID:   row.ID,
		Name: row.Name,
		Slug: row.Slug,
	}, nil
}

func (r *TaxonomyRepository) DeleteTag(ctx context.Context, slug string) error {
	return r.queries.DeleteTagBySlug(ctx, slug)
}
