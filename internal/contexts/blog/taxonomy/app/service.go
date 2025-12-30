package taxonomy

import (
	"context"
	"errors"
	"strings"

	taxdomain "proto-gin-web/internal/contexts/blog/taxonomy/domain"
)

// TaxonomyService coordinates category and tag operations.
type TaxonomyService interface {
	CreateCategory(ctx context.Context, input taxdomain.CreateCategoryInput) (taxdomain.Category, error)
	DeleteCategory(ctx context.Context, slug string) error
	CreateTag(ctx context.Context, input taxdomain.CreateTagInput) (taxdomain.Tag, error)
	DeleteTag(ctx context.Context, slug string) error
}

// Service implements TaxonomyService with validation and repository delegation.
type Service struct {
	repo taxdomain.TaxonomyRepository
}

var _ TaxonomyService = (*Service)(nil)

// NewService creates a taxonomy service.
func NewService(repo taxdomain.TaxonomyRepository) *Service {
	return &Service{repo: repo}
}

// CreateCategory validates input and persists a new category.
func (s *Service) CreateCategory(ctx context.Context, input taxdomain.CreateCategoryInput) (taxdomain.Category, error) {
	normalized, err := normalizeNameSlug(input.Name, input.Slug)
	if err != nil {
		return taxdomain.Category{}, err
	}
	return s.repo.CreateCategory(ctx, taxdomain.CreateCategoryInput{
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
func (s *Service) CreateTag(ctx context.Context, input taxdomain.CreateTagInput) (taxdomain.Tag, error) {
	normalized, err := normalizeNameSlug(input.Name, input.Slug)
	if err != nil {
		return taxdomain.Tag{}, err
	}
	return s.repo.CreateTag(ctx, taxdomain.CreateTagInput{
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

