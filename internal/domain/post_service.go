package domain

import (
	"context"
	"time"
)

// PostService exposes the application-facing operations around posts.
// It orchestrates domain rules while delegating persistence to repository implementations.
type PostService interface {
	ListPublished(ctx context.Context, opts ListPostsOptions) ([]Post, error)
	GetBySlug(ctx context.Context, slug string) (PostWithRelations, error)
	Create(ctx context.Context, input CreatePostInput) (Post, error)
	Update(ctx context.Context, input UpdatePostInput) (Post, error)
	Delete(ctx context.Context, slug string) error
	AddCategory(ctx context.Context, slug, categorySlug string) error
	RemoveCategory(ctx context.Context, slug, categorySlug string) error
	AddTag(ctx context.Context, slug, tagSlug string) error
	RemoveTag(ctx context.Context, slug, tagSlug string) error
}

// ListPostsOptions describes pagination and filtering instructions.
type ListPostsOptions struct {
	Category string
	Tag      string
	Sort     string
	Limit    int32
	Offset   int32
}

// PostWithRelations bundles a post along with its taxonomy associations.
type PostWithRelations struct {
	Post       Post       `json:"post"`
	Categories []Category `json:"categories"`
	Tags       []Tag      `json:"tags"`
}

// CreatePostInput describes the data required to create a post.
type CreatePostInput struct {
	Title       string
	Slug        string
	Summary     string
	ContentMD   string
	CoverURL    *string
	Status      string
	AuthorID    int64
	PublishedAt *time.Time
}

// UpdatePostInput captures editable fields for an existing post identified by slug.
type UpdatePostInput struct {
	Slug      string
	Title     string
	Summary   string
	ContentMD string
	CoverURL  *string
	Status    string
}
