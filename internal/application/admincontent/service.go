package admincontent

import (
	"context"
	"errors"
	"strings"

	"proto-gin-web/internal/domain"
)

// Service coordinates admin REST content operations.
type Service struct {
	posts    domain.PostService
	taxonomy domain.TaxonomyService
}

// NewService constructs a Service.
func NewService(posts domain.PostService, taxonomy domain.TaxonomyService) *Service {
	return &Service{posts: posts, taxonomy: taxonomy}
}

// CreatePost creates a post from API payload.
func (s *Service) CreatePost(ctx context.Context, input domain.CreatePostInput) (domain.Post, error) {
	normalizeCreate(&input)
	return s.posts.Create(ctx, input)
}

// UpdatePost updates a post by slug.
func (s *Service) UpdatePost(ctx context.Context, input domain.UpdatePostInput) (domain.Post, error) {
	normalizeUpdate(&input)
	return s.posts.Update(ctx, input)
}

// DeletePost removes a post.
func (s *Service) DeletePost(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return errors.New("admincontent: slug is required")
	}
	return s.posts.Delete(ctx, slug)
}

func (s *Service) AddCategory(ctx context.Context, slug, categorySlug string) error {
	return s.posts.AddCategory(ctx, strings.TrimSpace(slug), strings.TrimSpace(categorySlug))
}

func (s *Service) RemoveCategory(ctx context.Context, slug, categorySlug string) error {
	return s.posts.RemoveCategory(ctx, strings.TrimSpace(slug), strings.TrimSpace(categorySlug))
}

func (s *Service) AddTag(ctx context.Context, slug, tagSlug string) error {
	return s.posts.AddTag(ctx, strings.TrimSpace(slug), strings.TrimSpace(tagSlug))
}

func (s *Service) RemoveTag(ctx context.Context, slug, tagSlug string) error {
	return s.posts.RemoveTag(ctx, strings.TrimSpace(slug), strings.TrimSpace(tagSlug))
}

func (s *Service) CreateCategory(ctx context.Context, input domain.CreateCategoryInput) (domain.Category, error) {
	normalizeCat(&input)
	return s.taxonomy.CreateCategory(ctx, input)
}

func (s *Service) DeleteCategory(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return errors.New("admincontent: category slug is required")
	}
	return s.taxonomy.DeleteCategory(ctx, slug)
}

func (s *Service) CreateTag(ctx context.Context, input domain.CreateTagInput) (domain.Tag, error) {
	normalizeTag(&input)
	return s.taxonomy.CreateTag(ctx, input)
}

func (s *Service) DeleteTag(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return errors.New("admincontent: tag slug is required")
	}
	return s.taxonomy.DeleteTag(ctx, slug)
}

func normalizeCreate(input *domain.CreatePostInput) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Status = strings.TrimSpace(input.Status)
	if input.CoverURL != nil {
		trimmed := strings.TrimSpace(*input.CoverURL)
		input.CoverURL = &trimmed
	}
}

func normalizeUpdate(input *domain.UpdatePostInput) {
	input.Slug = strings.TrimSpace(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Status = strings.TrimSpace(input.Status)
	if input.CoverURL != nil {
		trimmed := strings.TrimSpace(*input.CoverURL)
		input.CoverURL = &trimmed
	}
}

func normalizeCat(input *domain.CreateCategoryInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
}

func normalizeTag(input *domain.CreateTagInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
}
