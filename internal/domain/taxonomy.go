package domain

import "context"

// CreateCategoryInput holds fields required to create a category.
type CreateCategoryInput struct {
	Name string
	Slug string
}

// CreateTagInput holds fields required to create a tag.
type CreateTagInput struct {
	Name string
	Slug string
}

// TaxonomyService coordinates category and tag operations.
type TaxonomyService interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, slug string) error
	CreateTag(ctx context.Context, input CreateTagInput) (Tag, error)
	DeleteTag(ctx context.Context, slug string) error
}

// TaxonomyRepository abstracts persistence of categories and tags.
type TaxonomyRepository interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, slug string) error
	CreateTag(ctx context.Context, input CreateTagInput) (Tag, error)
	DeleteTag(ctx context.Context, slug string) error
}
