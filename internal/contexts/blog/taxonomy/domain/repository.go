package taxdomain

import "context"

// TaxonomyRepository abstracts persistence of categories and tags.
type TaxonomyRepository interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, slug string) error
	CreateTag(ctx context.Context, input CreateTagInput) (Tag, error)
	DeleteTag(ctx context.Context, slug string) error
}
