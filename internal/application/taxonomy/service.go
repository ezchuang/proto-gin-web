package taxonomy

import (
	"context"
	"errors"
	"strings"

	"proto-gin-web/internal/domain"
)

// Service implements domain.TaxonomyService with validation and repository delegation.
type Service struct {
	repo domain.TaxonomyRepository
}

var _ domain.TaxonomyService = (*Service)(nil)

// NewService creates a taxonomy service.
func NewService(repo domain.TaxonomyRepository) *Service {
	return &Service{repo: repo}
}

// CreateCategory validates input and persists a new category.
func (s *Service) CreateCategory(ctx context.Context, input domain.CreateCategoryInput) (domain.Category, error) {
	normalized, err := normalizeNameSlug(input.Name, input.Slug)
	if err != nil {
		return domain.Category{}, err
	}
	return s.repo.CreateCategory(ctx, domain.CreateCategoryInput{
		Name: normalized.name,
		Slug: normalized.slug,
	})
}

// DeleteCategory removes a category by slug.
func (s *Service) DeleteCategory(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errors.New("taxonomy: category slug is required")
	}
	return s.repo.DeleteCategory(ctx, strings.TrimSpace(slug))
}

// CreateTag validates input and persists a new tag.
func (s *Service) CreateTag(ctx context.Context, input domain.CreateTagInput) (domain.Tag, error) {
	normalized, err := normalizeNameSlug(input.Name, input.Slug)
	if err != nil {
		return domain.Tag{}, err
	}
	return s.repo.CreateTag(ctx, domain.CreateTagInput{
		Name: normalized.name,
		Slug: normalized.slug,
	})
}

// DeleteTag removes a tag by slug.
func (s *Service) DeleteTag(ctx context.Context, slug string) error {
	if strings.TrimSpace(slug) == "" {
		return errors.New("taxonomy: tag slug is required")
	}
	return s.repo.DeleteTag(ctx, strings.TrimSpace(slug))
}

type normalizedPair struct {
	name string
	slug string
}

func normalizeNameSlug(name, slug string) (normalizedPair, error) {
	n := strings.TrimSpace(name)
	s := strings.TrimSpace(slug)
	if n == "" {
		return normalizedPair{}, errors.New("taxonomy: name is required")
	}
	if s == "" {
		return normalizedPair{}, errors.New("taxonomy: slug is required")
	}
	return normalizedPair{name: n, slug: strings.ToLower(s)}, nil
}
