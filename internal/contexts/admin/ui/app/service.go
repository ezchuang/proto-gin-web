package adminui

import (
	"context"
	"errors"
	"strings"

	postdomain "proto-gin-web/internal/contexts/blog/post/domain"
	postusecase "proto-gin-web/internal/contexts/blog/post/app"
)

// Service wraps post operations used by the admin UI forms.
type Service struct {
	posts postusecase.PostService
}

// NewService creates an admin UI helper service.
func NewService(posts postusecase.PostService) *Service {
	return &Service{posts: posts}
}

// ListPosts lists published posts for the UI with optional limit.
func (s *Service) ListPosts(ctx context.Context, limit int32) ([]postdomain.Post, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.posts.ListPublished(ctx, postdomain.ListPostsOptions{Limit: limit})
}

// GetPost fetches a post and its relations by slug.
func (s *Service) GetPost(ctx context.Context, slug string) (postdomain.PostWithRelations, error) {
	return s.posts.GetBySlug(ctx, strings.TrimSpace(slug))
}

// CreatePostParams captures the form fields required to create a post.
type CreatePostParams struct {
	Title     string
	Slug      string
	Summary   string
	ContentMD string
	CoverURL  string
	Status    string
	AuthorID  int64
}

// CreatePost creates a post from admin form params.
func (s *Service) CreatePost(ctx context.Context, params CreatePostParams) (postdomain.Post, error) {
	input := postdomain.CreatePostInput{
		Title:     strings.TrimSpace(params.Title),
		Slug:      strings.TrimSpace(params.Slug),
		Summary:   strings.TrimSpace(params.Summary),
		ContentMD: params.ContentMD,
		Status:    strings.TrimSpace(params.Status),
		AuthorID:  params.AuthorID,
	}
	if trimmed := strings.TrimSpace(params.CoverURL); trimmed != "" {
		input.CoverURL = &trimmed
	}
	return s.posts.Create(ctx, input)
}

// UpdatePostParams captures editable post fields.
type UpdatePostParams struct {
	Slug      string
	Title     string
	Summary   string
	ContentMD string
	CoverURL  string
	Status    string
}

// UpdatePost updates a post identified by slug.
func (s *Service) UpdatePost(ctx context.Context, params UpdatePostParams) (postdomain.Post, error) {
	input := postdomain.UpdatePostInput{
		Slug:      strings.TrimSpace(params.Slug),
		Title:     strings.TrimSpace(params.Title),
		Summary:   strings.TrimSpace(params.Summary),
		ContentMD: params.ContentMD,
		Status:    strings.TrimSpace(params.Status),
	}
	if trimmed := strings.TrimSpace(params.CoverURL); trimmed != "" {
		input.CoverURL = &trimmed
	}
	return s.posts.Update(ctx, input)
}

// DeletePost removes a post by slug.
func (s *Service) DeletePost(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errors.New("adminui: slug is required")
	}
	return s.posts.Delete(ctx, strings.TrimSpace(slug))
}

// AddCategory assigns a category to a post.
func (s *Service) AddCategory(ctx context.Context, slug, categorySlug string) error {
	return s.posts.AddCategory(ctx, slug, categorySlug)
}

// RemoveCategory removes a category from a post.
func (s *Service) RemoveCategory(ctx context.Context, slug, categorySlug string) error {
	return s.posts.RemoveCategory(ctx, slug, categorySlug)
}

// AddTag assigns a tag to a post.
func (s *Service) AddTag(ctx context.Context, slug, tagSlug string) error {
	return s.posts.AddTag(ctx, slug, tagSlug)
}

// RemoveTag removes a tag from a post.
func (s *Service) RemoveTag(ctx context.Context, slug, tagSlug string) error {
	return s.posts.RemoveTag(ctx, slug, tagSlug)
}


