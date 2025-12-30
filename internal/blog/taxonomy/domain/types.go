package taxdomain

import "context"

// Category represents a taxonomy group assigned to posts.
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Tag is a flexible label attached to posts.
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

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

// TaxonomyRepository abstracts persistence of categories and tags.
type TaxonomyRepository interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, slug string) error
	CreateTag(ctx context.Context, input CreateTagInput) (Tag, error)
	DeleteTag(ctx context.Context, slug string) error
}
