package taxonomy

import (
	"context"
	"errors"
	"testing"

	taxdomain "proto-gin-web/internal/blog/taxonomy/domain"
)

type mockRepo struct {
	categoryInput taxdomain.CreateCategoryInput
	tagInput      taxdomain.CreateTagInput
	categorySlug  string
	tagSlug       string

	categoryResult taxdomain.Category
	tagResult      taxdomain.Tag

	errCategory error
	errTag      error
	errDelete   error
}

func (m *mockRepo) CreateCategory(ctx context.Context, input taxdomain.CreateCategoryInput) (taxdomain.Category, error) {
	m.categoryInput = input
	return m.categoryResult, m.errCategory
}

func (m *mockRepo) DeleteCategory(ctx context.Context, slug string) error {
	m.categorySlug = slug
	return m.errDelete
}

func (m *mockRepo) CreateTag(ctx context.Context, input taxdomain.CreateTagInput) (taxdomain.Tag, error) {
	m.tagInput = input
	return m.tagResult, m.errTag
}

func (m *mockRepo) DeleteTag(ctx context.Context, slug string) error {
	m.tagSlug = slug
	return m.errDelete
}

func TestService_CreateCategory_normalizesInput(t *testing.T) {
	repo := &mockRepo{
		categoryResult: taxdomain.Category{ID: 1, Name: "Foo", Slug: "foo"},
	}
	svc := NewService(repo)

	result, err := svc.CreateCategory(context.Background(), taxdomain.CreateCategoryInput{
		Name: "  Foo  ",
		Slug: "  FOO ",
	})
	if err != nil {
		t.Fatalf("CreateCategory returned error: %v", err)
	}
	if result != (taxdomain.Category{ID: 1, Name: "Foo", Slug: "foo"}) {
		t.Fatalf("unexpected category result: %+v", result)
	}
	if repo.categoryInput.Name != "Foo" || repo.categoryInput.Slug != "foo" {
		t.Fatalf("input not normalized: %+v", repo.categoryInput)
	}
}

func TestService_CreateCategory_validation(t *testing.T) {
	svc := NewService(&mockRepo{})

	_, err := svc.CreateCategory(context.Background(), taxdomain.CreateCategoryInput{
		Name: "",
		Slug: "slug",
	})
	if err == nil || err.Error() != "taxonomy: name is required" {
		t.Fatalf("expected name error, got %v", err)
	}

	_, err = svc.CreateCategory(context.Background(), taxdomain.CreateCategoryInput{
		Name: "Name",
		Slug: "",
	})
	if err == nil || err.Error() != "taxonomy: slug is required" {
		t.Fatalf("expected slug error, got %v", err)
	}
}

func TestService_DeleteCategory(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	if err := svc.DeleteCategory(context.Background(), "  slug  "); err != nil {
		t.Fatalf("DeleteCategory returned error: %v", err)
	}
	if repo.categorySlug != "slug" {
		t.Fatalf("unexpected slug stored: %s", repo.categorySlug)
	}

	err := svc.DeleteCategory(context.Background(), " ")
	if err == nil || err.Error() != "taxonomy: category slug is required" {
		t.Fatalf("expected slug required error, got %v", err)
	}
}

func TestService_CreateTag_normalizesInput(t *testing.T) {
	repo := &mockRepo{
		tagResult: taxdomain.Tag{ID: 1, Name: "Bar", Slug: "bar"},
	}
	svc := NewService(repo)

	result, err := svc.CreateTag(context.Background(), taxdomain.CreateTagInput{
		Name: "  Bar ",
		Slug: "  BAR ",
	})
	if err != nil {
		t.Fatalf("CreateTag returned error: %v", err)
	}
	if result != (taxdomain.Tag{ID: 1, Name: "Bar", Slug: "bar"}) {
		t.Fatalf("unexpected tag result: %+v", result)
	}
	if repo.tagInput.Name != "Bar" || repo.tagInput.Slug != "bar" {
		t.Fatalf("input not normalized: %+v", repo.tagInput)
	}
}

func TestService_CreateTag_validation(t *testing.T) {
	svc := NewService(&mockRepo{})

	_, err := svc.CreateTag(context.Background(), taxdomain.CreateTagInput{
		Name: "",
		Slug: "slug",
	})
	if err == nil || err.Error() != "taxonomy: name is required" {
		t.Fatalf("expected name error, got %v", err)
	}

	_, err = svc.CreateTag(context.Background(), taxdomain.CreateTagInput{
		Name: "Name",
		Slug: "",
	})
	if err == nil || err.Error() != "taxonomy: slug is required" {
		t.Fatalf("expected slug error, got %v", err)
	}
}

func TestService_DeleteTag(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	if err := svc.DeleteTag(context.Background(), "  slug "); err != nil {
		t.Fatalf("DeleteTag returned error: %v", err)
	}
	if repo.tagSlug != "slug" {
		t.Fatalf("unexpected slug stored: %s", repo.tagSlug)
	}

	err := svc.DeleteTag(context.Background(), "")
	if err == nil || err.Error() != "taxonomy: tag slug is required" {
		t.Fatalf("expected slug required error, got %v", err)
	}
}

func TestService_RepoErrorsPropagated(t *testing.T) {
	repo := &mockRepo{
		errCategory: errors.New("db error"),
		errTag:      errors.New("db error"),
		errDelete:   errors.New("delete error"),
	}
	svc := NewService(repo)

	_, err := svc.CreateCategory(context.Background(), taxdomain.CreateCategoryInput{Name: "Foo", Slug: "foo"})
	if err == nil || err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", err)
	}

	_, err = svc.CreateTag(context.Background(), taxdomain.CreateTagInput{Name: "Bar", Slug: "bar"})
	if err == nil || err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", err)
	}

	err = svc.DeleteCategory(context.Background(), "slug")
	if err == nil || err.Error() != "delete error" {
		t.Fatalf("expected delete error, got %v", err)
	}

	err = svc.DeleteTag(context.Background(), "slug")
	if err == nil || err.Error() != "delete error" {
		t.Fatalf("expected delete error, got %v", err)
	}
}
