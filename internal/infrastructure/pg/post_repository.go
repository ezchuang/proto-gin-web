package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	postdomain "proto-gin-web/internal/blog/post/domain"
	taxdomain "proto-gin-web/internal/blog/taxonomy/domain"
)

// PostRepository provides a postdomain.PostRepository backed by pgx queries.
type PostRepository struct {
	queries *Queries
}

// NewPostRepository constructs a PostRepository from a pool.
func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{queries: New(pool)}
}

var _ postdomain.PostRepository = (*PostRepository)(nil)

func (r *PostRepository) ListPublishedPosts(ctx context.Context, limit, offset int32) ([]postdomain.Post, error) {
	posts, err := r.queries.ListPublishedPosts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapPosts(posts), nil
}

func (r *PostRepository) ListPublishedPostsSorted(ctx context.Context, sort string, limit, offset int32) ([]postdomain.Post, error) {
	posts, err := r.queries.ListPublishedPostsSorted(ctx, sort, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapPosts(posts), nil
}

func (r *PostRepository) ListPublishedPostsByCategorySorted(ctx context.Context, categorySlug, sort string, limit, offset int32) ([]postdomain.Post, error) {
	posts, err := r.queries.ListPublishedPostsByCategorySorted(ctx, categorySlug, sort, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapPosts(posts), nil
}

func (r *PostRepository) ListPublishedPostsByTagSorted(ctx context.Context, tagSlug, sort string, limit, offset int32) ([]postdomain.Post, error) {
	posts, err := r.queries.ListPublishedPostsByTagSorted(ctx, tagSlug, sort, limit, offset)
	if err != nil {
		return nil, err
	}
	return mapPosts(posts), nil
}

func (r *PostRepository) GetPostBySlug(ctx context.Context, slug string) (postdomain.Post, error) {
	post, err := r.queries.GetPostBySlug(ctx, slug)
	if err != nil {
		return postdomain.Post{}, err
	}
	return mapPost(post), nil
}

func (r *PostRepository) CreatePost(ctx context.Context, input postdomain.CreatePostInput) (postdomain.Post, error) {
	params := CreatePostParams{
		Title:       input.Title,
		Slug:        input.Slug,
		Summary:     input.Summary,
		ContentMd:   input.ContentMD,
		Status:      input.Status,
		AuthorID:    input.AuthorID,
		PublishedAt: input.PublishedAt,
	}
	if input.CoverURL != nil {
		params.CoverUrl = input.CoverURL
	}
	post, err := r.queries.CreatePost(ctx, params)
	if err != nil {
		return postdomain.Post{}, err
	}
	return mapPost(post), nil
}

func (r *PostRepository) UpdatePostBySlug(ctx context.Context, input postdomain.UpdatePostInput) (postdomain.Post, error) {
	params := UpdatePostBySlugParams{
		Slug:      input.Slug,
		Title:     input.Title,
		Summary:   input.Summary,
		ContentMd: input.ContentMD,
		Status:    input.Status,
	}
	if input.CoverURL != nil {
		params.CoverUrl = input.CoverURL
	}
	post, err := r.queries.UpdatePostBySlug(ctx, params)
	if err != nil {
		return postdomain.Post{}, err
	}
	return mapPost(post), nil
}

func (r *PostRepository) DeletePostBySlug(ctx context.Context, slug string) error {
	return r.queries.DeletePostBySlug(ctx, slug)
}

func (r *PostRepository) AddCategoryToPost(ctx context.Context, slug, categorySlug string) error {
	return r.queries.AddCategoryToPost(ctx, slug, categorySlug)
}

func (r *PostRepository) RemoveCategoryFromPost(ctx context.Context, slug, categorySlug string) error {
	return r.queries.RemoveCategoryFromPost(ctx, slug, categorySlug)
}

func (r *PostRepository) AddTagToPost(ctx context.Context, slug, tagSlug string) error {
	return r.queries.AddTagToPost(ctx, slug, tagSlug)
}

func (r *PostRepository) RemoveTagFromPost(ctx context.Context, slug, tagSlug string) error {
	return r.queries.RemoveTagFromPost(ctx, slug, tagSlug)
}

func (r *PostRepository) ListCategoriesByPostSlug(ctx context.Context, slug string) ([]taxdomain.Category, error) {
	cats, err := r.queries.ListCategoriesByPostSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return mapCategories(cats), nil
}

func (r *PostRepository) ListTagsByPostSlug(ctx context.Context, slug string) ([]taxdomain.Tag, error) {
	tags, err := r.queries.ListTagsByPostSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return mapTags(tags), nil
}

func mapPosts(posts []Post) []postdomain.Post {
	out := make([]postdomain.Post, len(posts))
	for i, p := range posts {
		out[i] = mapPost(p)
	}
	return out
}

func mapPost(p Post) postdomain.Post {
	return postdomain.Post{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Summary:     p.Summary,
		ContentMD:   p.ContentMd,
		CoverURL:    p.CoverUrl,
		Status:      p.Status,
		AuthorID:    p.AuthorID,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func mapCategories(categories []Category) []taxdomain.Category {
	out := make([]taxdomain.Category, len(categories))
	for i, c := range categories {
		out[i] = taxdomain.Category{ID: c.ID, Name: c.Name, Slug: c.Slug}
	}
	return out
}

func mapTags(tags []Tag) []taxdomain.Tag {
	out := make([]taxdomain.Tag, len(tags))
	for i, t := range tags {
		out[i] = taxdomain.Tag{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}
	return out
}
