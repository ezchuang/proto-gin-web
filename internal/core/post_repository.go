package core

import "context"

// PostRepository abstracts persistence operations for posts and their relations.
type PostRepository interface {
	ListPublishedPosts(ctx context.Context, limit, offset int32) ([]Post, error)
	ListPublishedPostsSorted(ctx context.Context, sort string, limit, offset int32) ([]Post, error)
	ListPublishedPostsByCategorySorted(ctx context.Context, categorySlug, sort string, limit, offset int32) ([]Post, error)
	ListPublishedPostsByTagSorted(ctx context.Context, tagSlug, sort string, limit, offset int32) ([]Post, error)

	GetPostBySlug(ctx context.Context, slug string) (Post, error)
	CreatePost(ctx context.Context, input CreatePostInput) (Post, error)
	UpdatePostBySlug(ctx context.Context, input UpdatePostInput) (Post, error)
	DeletePostBySlug(ctx context.Context, slug string) error

	AddCategoryToPost(ctx context.Context, slug, categorySlug string) error
	RemoveCategoryFromPost(ctx context.Context, slug, categorySlug string) error
	AddTagToPost(ctx context.Context, slug, tagSlug string) error
	RemoveTagFromPost(ctx context.Context, slug, tagSlug string) error

	ListCategoriesByPostSlug(ctx context.Context, slug string) ([]Category, error)
	ListTagsByPostSlug(ctx context.Context, slug string) ([]Tag, error)
}
