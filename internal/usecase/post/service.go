package post

import (
	"context"
	"errors"
	"strings"

	"proto-gin-web/internal/core"
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

// Service implements core.PostService using a repository abstraction.
type Service struct {
	repo core.PostRepository
}

var _ core.PostService = (*Service)(nil)

// NewService wires a post repository into a use case implementation.
func NewService(repo core.PostRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPublished(ctx context.Context, opts core.ListPostsOptions) ([]core.Post, error) {
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

func (s *Service) GetBySlug(ctx context.Context, slug string) (core.PostWithRelations, error) {
	if strings.TrimSpace(slug) == "" {
		return core.PostWithRelations{}, errSlugRequired
	}

	post, err := s.repo.GetPostBySlug(ctx, slug)
	if err != nil {
		return core.PostWithRelations{}, err
	}

	cats, err := s.repo.ListCategoriesByPostSlug(ctx, slug)
	if err != nil {
		return core.PostWithRelations{}, err
	}
	tags, err := s.repo.ListTagsByPostSlug(ctx, slug)
	if err != nil {
		return core.PostWithRelations{}, err
	}

	return core.PostWithRelations{Post: post, Categories: cats, Tags: tags}, nil
}

func (s *Service) Create(ctx context.Context, input core.CreatePostInput) (core.Post, error) {
	if err := validateCreateInput(input); err != nil {
		return core.Post{}, err
	}
	return s.repo.CreatePost(ctx, normalizeCreateInput(input))
}

func (s *Service) Update(ctx context.Context, input core.UpdatePostInput) (core.Post, error) {
	if err := validateUpdateInput(input); err != nil {
		return core.Post{}, err
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

func validateCreateInput(input core.CreatePostInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return errTitleRequired
	}
	if strings.TrimSpace(input.Slug) == "" {
		return errSlugRequired
	}
	return nil
}

func validateUpdateInput(input core.UpdatePostInput) error {
	if strings.TrimSpace(input.Slug) == "" {
		return errSlugRequired
	}
	if strings.TrimSpace(input.Title) == "" {
		return errTitleRequired
	}
	return nil
}

func normalizeCreateInput(input core.CreatePostInput) core.CreatePostInput {
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

func normalizeUpdateInput(input core.UpdatePostInput) core.UpdatePostInput {
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
