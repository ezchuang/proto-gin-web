package post

import (
	"context"
	"errors"
	"strings"

	"proto-gin-web/internal/domain"
)

const (
	defaultPageSize int32 = 10
	maxPageSize     int32 = 50
)

var (
	errTitleRequired = errors.New("title is required")
	errSlugRequired  = errors.New("slug is required")
)

var allowedSorts = map[string]struct{}{
	"created_at_desc":   {},
	"created_at_asc":    {},
	"published_at_desc": {},
	"published_at_asc":  {},
	"":                  {},
}

// Service implements domain.PostService using a repository abstraction.
type Service struct {
	repo domain.PostRepository
}

var _ domain.PostService = (*Service)(nil)

// NewService wires a post repository into a use case implementation.
func NewService(repo domain.PostRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPublished(ctx context.Context, opts domain.ListPostsOptions) ([]domain.Post, error) {
	limit := clampLimit(opts.Limit)
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	sort := normalizeSort(opts.Sort)

	switch {
	case opts.Category != "":
		return s.repo.ListPublishedPostsByCategorySorted(ctx, opts.Category, sort, limit, offset)
	case opts.Tag != "":
		return s.repo.ListPublishedPostsByTagSorted(ctx, opts.Tag, sort, limit, offset)
	default:
		return s.repo.ListPublishedPostsSorted(ctx, sort, limit, offset)
	}
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (domain.PostWithRelations, error) {
	if strings.TrimSpace(slug) == "" {
		return domain.PostWithRelations{}, errSlugRequired
	}

	post, err := s.repo.GetPostBySlug(ctx, slug)
	if err != nil {
		return domain.PostWithRelations{}, err
	}

	cats, err := s.repo.ListCategoriesByPostSlug(ctx, slug)
	if err != nil {
		return domain.PostWithRelations{}, err
	}
	tags, err := s.repo.ListTagsByPostSlug(ctx, slug)
	if err != nil {
		return domain.PostWithRelations{}, err
	}

	return domain.PostWithRelations{Post: post, Categories: cats, Tags: tags}, nil
}

func (s *Service) Create(ctx context.Context, input domain.CreatePostInput) (domain.Post, error) {
	if err := validateCreateInput(input); err != nil {
		return domain.Post{}, err
	}
	return s.repo.CreatePost(ctx, normalizeCreateInput(input))
}

func (s *Service) Update(ctx context.Context, input domain.UpdatePostInput) (domain.Post, error) {
	if err := validateUpdateInput(input); err != nil {
		return domain.Post{}, err
	}
	return s.repo.UpdatePostBySlug(ctx, normalizeUpdateInput(input))
}

func (s *Service) Delete(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errSlugRequired
	}
	return s.repo.DeletePostBySlug(ctx, slug)
}

func (s *Service) AddCategory(ctx context.Context, slug, categorySlug string) error {
	if strings.TrimSpace(slug) == "" {
		return errSlugRequired
	}
	if strings.TrimSpace(categorySlug) == "" {
		return errors.New("category slug is required")
	}
	return s.repo.AddCategoryToPost(ctx, slug, categorySlug)
}

func (s *Service) RemoveCategory(ctx context.Context, slug, categorySlug string) error {
	if strings.TrimSpace(slug) == "" {
		return errSlugRequired
	}
	if strings.TrimSpace(categorySlug) == "" {
		return errors.New("category slug is required")
	}
	return s.repo.RemoveCategoryFromPost(ctx, slug, categorySlug)
}

func (s *Service) AddTag(ctx context.Context, slug, tagSlug string) error {
	if strings.TrimSpace(slug) == "" {
		return errSlugRequired
	}
	if strings.TrimSpace(tagSlug) == "" {
		return errors.New("tag slug is required")
	}
	return s.repo.AddTagToPost(ctx, slug, tagSlug)
}

func (s *Service) RemoveTag(ctx context.Context, slug, tagSlug string) error {
	if strings.TrimSpace(slug) == "" {
		return errSlugRequired
	}
	if strings.TrimSpace(tagSlug) == "" {
		return errors.New("tag slug is required")
	}
	return s.repo.RemoveTagFromPost(ctx, slug, tagSlug)
}

func clampLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultPageSize
	}
	if limit > maxPageSize {
		return maxPageSize
	}
	return limit
}

func normalizeSort(sort string) string {
	sort = strings.ToLower(strings.TrimSpace(sort))
	if _, ok := allowedSorts[sort]; ok {
		return sort
	}
	return "created_at_desc"
}

func validateCreateInput(input domain.CreatePostInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return errTitleRequired
	}
	if strings.TrimSpace(input.Slug) == "" {
		return errSlugRequired
	}
	return nil
}

func validateUpdateInput(input domain.UpdatePostInput) error {
	if strings.TrimSpace(input.Slug) == "" {
		return errSlugRequired
	}
	if strings.TrimSpace(input.Title) == "" {
		return errTitleRequired
	}
	return nil
}

func normalizeCreateInput(input domain.CreatePostInput) domain.CreatePostInput {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Status = strings.TrimSpace(input.Status)
	input.Summary = strings.TrimSpace(input.Summary)
	if input.CoverURL != nil {
		trimmed := strings.TrimSpace(*input.CoverURL)
		if trimmed == "" {
			input.CoverURL = nil
		} else {
			input.CoverURL = &trimmed
		}
	}
	return input
}

func normalizeUpdateInput(input domain.UpdatePostInput) domain.UpdatePostInput {
	input.Slug = strings.TrimSpace(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Status = strings.TrimSpace(input.Status)
	if input.CoverURL != nil {
		trimmed := strings.TrimSpace(*input.CoverURL)
		if trimmed == "" {
			input.CoverURL = nil
		} else {
			input.CoverURL = &trimmed
		}
	}
	return input
}
