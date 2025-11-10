package admincontent

import (
	"context"
	"errors"
	"testing"

	"proto-gin-web/internal/domain"
)

func TestService_CreateAndUpdatePost_normalizesInput(t *testing.T) {
	postSvc := &stubPostSvc{
		createResult: domain.Post{ID: 1, Slug: "hello-world"},
		updateResult: domain.Post{ID: 1, Slug: "hello-world"},
	}
	svc := NewService(postSvc, &stubTaxonomySvc{})

	cover := "  https://example.com/image.jpg  "
	if _, err := svc.CreatePost(context.Background(), domain.CreatePostInput{
		Title:     "  Title  ",
		Slug:      "  hello-world  ",
		Summary:   " summary ",
		ContentMD: "content",
		CoverURL:  &cover,
		Status:    " draft ",
	}); err != nil {
		t.Fatalf("CreatePost returned error: %v", err)
	}
	if postSvc.createInput.Title != "Title" ||
		postSvc.createInput.Slug != "hello-world" ||
		postSvc.createInput.Summary != "summary" ||
		postSvc.createInput.Status != "draft" {
		t.Fatalf("CreatePost input not normalized: %+v", postSvc.createInput)
	}
	if postSvc.createInput.CoverURL == nil || *postSvc.createInput.CoverURL != "https://example.com/image.jpg" {
		t.Fatalf("expected trimmed cover url, got %v", postSvc.createInput.CoverURL)
	}

	updateCover := "  /cover.png "
	if _, err := svc.UpdatePost(context.Background(), domain.UpdatePostInput{
		Slug:      "  hello-world  ",
		Title:     "  Updated ",
		Summary:   " summary ",
		ContentMD: "content",
		Status:    " published ",
		CoverURL:  &updateCover,
	}); err != nil {
		t.Fatalf("UpdatePost returned error: %v", err)
	}
	if postSvc.updateInput.Slug != "hello-world" ||
		postSvc.updateInput.Title != "Updated" ||
		postSvc.updateInput.Status != "published" {
		t.Fatalf("UpdatePost input not normalized: %+v", postSvc.updateInput)
	}
	if postSvc.updateInput.CoverURL == nil || *postSvc.updateInput.CoverURL != "/cover.png" {
		t.Fatalf("expected trimmed cover url on update, got %v", postSvc.updateInput.CoverURL)
	}
}

func TestService_DeletePost_validatesSlug(t *testing.T) {
	postSvc := &stubPostSvc{}
	svc := NewService(postSvc, &stubTaxonomySvc{})

	if err := svc.DeletePost(context.Background(), "  hello-world  "); err != nil {
		t.Fatalf("DeletePost returned error: %v", err)
	}
	if postSvc.deleteSlug != "hello-world" {
		t.Fatalf("expected trimmed slug, got %q", postSvc.deleteSlug)
	}

	if err := svc.DeletePost(context.Background(), "   "); err == nil || err.Error() != "admincontent: slug is required" {
		t.Fatalf("expected slug required error, got %v", err)
	}
}

func TestService_TaxonomyOperations_normalizeInput(t *testing.T) {
	taxSvc := &stubTaxonomySvc{
		categoryResult: domain.Category{ID: 1, Name: "Foo", Slug: "foo"},
		tagResult:      domain.Tag{ID: 2, Name: "Bar", Slug: "bar"},
	}
	svc := NewService(&stubPostSvc{}, taxSvc)

	if _, err := svc.CreateCategory(context.Background(), domain.CreateCategoryInput{
		Name: "  Foo ",
		Slug: "  FOO ",
	}); err != nil {
		t.Fatalf("CreateCategory returned error: %v", err)
	}
	if taxSvc.categoryInput.Name != "Foo" || taxSvc.categoryInput.Slug != "foo" {
		t.Fatalf("category input not normalized: %+v", taxSvc.categoryInput)
	}

	if err := svc.DeleteCategory(context.Background(), "  foo "); err != nil {
		t.Fatalf("DeleteCategory returned error: %v", err)
	}
	if taxSvc.categorySlug != "foo" {
		t.Fatalf("expected trimmed category slug, got %q", taxSvc.categorySlug)
	}
	if err := svc.DeleteCategory(context.Background(), " "); err == nil || err.Error() != "admincontent: category slug is required" {
		t.Fatalf("expected category slug required error, got %v", err)
	}

	if _, err := svc.CreateTag(context.Background(), domain.CreateTagInput{
		Name: "  Bar ",
		Slug: "  BAR ",
	}); err != nil {
		t.Fatalf("CreateTag returned error: %v", err)
	}
	if taxSvc.tagInput.Name != "Bar" || taxSvc.tagInput.Slug != "bar" {
		t.Fatalf("tag input not normalized: %+v", taxSvc.tagInput)
	}

	if err := svc.DeleteTag(context.Background(), "  bar "); err != nil {
		t.Fatalf("DeleteTag returned error: %v", err)
	}
	if taxSvc.tagSlug != "bar" {
		t.Fatalf("expected trimmed tag slug, got %q", taxSvc.tagSlug)
	}
	if err := svc.DeleteTag(context.Background(), " "); err == nil || err.Error() != "admincontent: tag slug is required" {
		t.Fatalf("expected tag slug required error, got %v", err)
	}
}

func TestService_CategoryAndTagAssignments_trimSlugs(t *testing.T) {
	postSvc := &stubPostSvc{}
	svc := NewService(postSvc, &stubTaxonomySvc{})

	if err := svc.AddCategory(context.Background(), "  post-slug  ", "  category "); err != nil {
		t.Fatalf("AddCategory returned error: %v", err)
	}
	if postSvc.addCategoryArgs != ([2]string{"post-slug", "category"}) {
		t.Fatalf("expected trimmed args, got %+v", postSvc.addCategoryArgs)
	}

	if err := svc.RemoveCategory(context.Background(), "  post-slug  ", "  category "); err != nil {
		t.Fatalf("RemoveCategory returned error: %v", err)
	}
	if postSvc.removeCategoryArgs != ([2]string{"post-slug", "category"}) {
		t.Fatalf("expected trimmed args, got %+v", postSvc.removeCategoryArgs)
	}

	if err := svc.AddTag(context.Background(), "  post-slug  ", "  tag "); err != nil {
		t.Fatalf("AddTag returned error: %v", err)
	}
	if postSvc.addTagArgs != ([2]string{"post-slug", "tag"}) {
		t.Fatalf("expected trimmed args, got %+v", postSvc.addTagArgs)
	}

	if err := svc.RemoveTag(context.Background(), "  post-slug  ", "  tag "); err != nil {
		t.Fatalf("RemoveTag returned error: %v", err)
	}
	if postSvc.removeTagArgs != ([2]string{"post-slug", "tag"}) {
		t.Fatalf("expected trimmed args, got %+v", postSvc.removeTagArgs)
	}
}

func TestService_ErrorPropagation(t *testing.T) {
	postSvc := &stubPostSvc{
		errCreate:      errors.New("create failed"),
		errUpdate:      errors.New("update failed"),
		errDelete:      errors.New("delete failed"),
		errAddCategory: errors.New("add cat failed"),
		errRemoveCat:   errors.New("remove cat failed"),
		errAddTag:      errors.New("add tag failed"),
		errRemoveTag:   errors.New("remove tag failed"),
	}
	taxSvc := &stubTaxonomySvc{
		errCreateCategory: errors.New("create category failed"),
		errDeleteCategory: errors.New("delete category failed"),
		errCreateTag:      errors.New("create tag failed"),
		errDeleteTag:      errors.New("delete tag failed"),
	}
	svc := NewService(postSvc, taxSvc)

	cover := "cover"
	if _, err := svc.CreatePost(context.Background(), domain.CreatePostInput{
		Title:     "t",
		Slug:      "s",
		ContentMD: "c",
		Status:    "draft",
		CoverURL:  &cover,
	}); err == nil || err.Error() != "create failed" {
		t.Fatalf("expected create error, got %v", err)
	}

	if _, err := svc.UpdatePost(context.Background(), domain.UpdatePostInput{
		Slug:      "s",
		Title:     "t",
		ContentMD: "c",
		Status:    "draft",
	}); err == nil || err.Error() != "update failed" {
		t.Fatalf("expected update error, got %v", err)
	}

	if err := svc.DeletePost(context.Background(), "slug"); err == nil || err.Error() != "delete failed" {
		t.Fatalf("expected delete error, got %v", err)
	}

	if err := svc.AddCategory(context.Background(), "slug", "cat"); err == nil || err.Error() != "add cat failed" {
		t.Fatalf("expected add category error, got %v", err)
	}

	if err := svc.RemoveCategory(context.Background(), "slug", "cat"); err == nil || err.Error() != "remove cat failed" {
		t.Fatalf("expected remove category error, got %v", err)
	}

	if err := svc.AddTag(context.Background(), "slug", "tag"); err == nil || err.Error() != "add tag failed" {
		t.Fatalf("expected add tag error, got %v", err)
	}

	if err := svc.RemoveTag(context.Background(), "slug", "tag"); err == nil || err.Error() != "remove tag failed" {
		t.Fatalf("expected remove tag error, got %v", err)
	}

	if _, err := svc.CreateCategory(context.Background(), domain.CreateCategoryInput{Name: "Foo", Slug: "foo"}); err == nil || err.Error() != "create category failed" {
		t.Fatalf("expected create category error, got %v", err)
	}

	if err := svc.DeleteCategory(context.Background(), "foo"); err == nil || err.Error() != "delete category failed" {
		t.Fatalf("expected delete category error, got %v", err)
	}

	if _, err := svc.CreateTag(context.Background(), domain.CreateTagInput{Name: "Bar", Slug: "bar"}); err == nil || err.Error() != "create tag failed" {
		t.Fatalf("expected create tag error, got %v", err)
	}

	if err := svc.DeleteTag(context.Background(), "bar"); err == nil || err.Error() != "delete tag failed" {
		t.Fatalf("expected delete tag error, got %v", err)
	}
}

type stubPostSvc struct {
	createInput domain.CreatePostInput
	updateInput domain.UpdatePostInput
	deleteSlug  string

	addCategoryArgs    [2]string
	removeCategoryArgs [2]string
	addTagArgs         [2]string
	removeTagArgs      [2]string

	createResult domain.Post
	updateResult domain.Post

	errCreate      error
	errUpdate      error
	errDelete      error
	errAddCategory error
	errRemoveCat   error
	errAddTag      error
	errRemoveTag   error
}

func (s *stubPostSvc) ListPublished(context.Context, domain.ListPostsOptions) ([]domain.Post, error) {
	return nil, nil
}

func (s *stubPostSvc) GetBySlug(context.Context, string) (domain.PostWithRelations, error) {
	return domain.PostWithRelations{}, nil
}

func (s *stubPostSvc) Create(ctx context.Context, input domain.CreatePostInput) (domain.Post, error) {
	s.createInput = input
	return s.createResult, s.errCreate
}

func (s *stubPostSvc) Update(ctx context.Context, input domain.UpdatePostInput) (domain.Post, error) {
	s.updateInput = input
	return s.updateResult, s.errUpdate
}

func (s *stubPostSvc) Delete(ctx context.Context, slug string) error {
	s.deleteSlug = slug
	return s.errDelete
}

func (s *stubPostSvc) AddCategory(ctx context.Context, slug, categorySlug string) error {
	s.addCategoryArgs = [2]string{slug, categorySlug}
	return s.errAddCategory
}

func (s *stubPostSvc) RemoveCategory(ctx context.Context, slug, categorySlug string) error {
	s.removeCategoryArgs = [2]string{slug, categorySlug}
	return s.errRemoveCat
}

func (s *stubPostSvc) AddTag(ctx context.Context, slug, tagSlug string) error {
	s.addTagArgs = [2]string{slug, tagSlug}
	return s.errAddTag
}

func (s *stubPostSvc) RemoveTag(ctx context.Context, slug, tagSlug string) error {
	s.removeTagArgs = [2]string{slug, tagSlug}
	return s.errRemoveTag
}

type stubTaxonomySvc struct {
	categoryInput domain.CreateCategoryInput
	tagInput      domain.CreateTagInput
	categorySlug  string
	tagSlug       string

	categoryResult domain.Category
	tagResult      domain.Tag

	errCreateCategory error
	errDeleteCategory error
	errCreateTag      error
	errDeleteTag      error
}

func (s *stubTaxonomySvc) CreateCategory(ctx context.Context, input domain.CreateCategoryInput) (domain.Category, error) {
	s.categoryInput = input
	return s.categoryResult, s.errCreateCategory
}

func (s *stubTaxonomySvc) DeleteCategory(ctx context.Context, slug string) error {
	s.categorySlug = slug
	return s.errDeleteCategory
}

func (s *stubTaxonomySvc) CreateTag(ctx context.Context, input domain.CreateTagInput) (domain.Tag, error) {
	s.tagInput = input
	return s.tagResult, s.errCreateTag
}

func (s *stubTaxonomySvc) DeleteTag(ctx context.Context, slug string) error {
	s.tagSlug = slug
	return s.errDeleteTag
}
