package post

import (
	"context"
	"errors"
	"testing"

	"proto-gin-web/internal/core"
)

func TestServiceListPublished_Defaults(t *testing.T) {
	repo := &fakePostRepo{}
	repo.listPublishedPostsSortedFn = func(ctx context.Context, sort string, limit, offset int32) ([]core.Post, error) {
		if sort != "" {
			t.Fatalf("expected empty sort passthrough, got %q", sort)
		}
		if limit != 10 { // defaultPageSize
			t.Fatalf("expected default limit 10, got %d", limit)
		}
		if offset != 0 {
			t.Fatalf("expected offset 0, got %d", offset)
		}
		return []core.Post{{Slug: "hello"}}, nil
	}

	svc := NewService(repo)
	result, err := svc.ListPublished(context.Background(), core.ListPostsOptions{})
	if err != nil {
		t.Fatalf("ListPublished returned error: %v", err)
	}
	if len(result) != 1 || result[0].Slug != "hello" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestServiceListPublished_Category(t *testing.T) {
	repo := &fakePostRepo{}
	repo.listPublishedPostsByCategorySortedFn = func(ctx context.Context, categorySlug, sort string, limit, offset int32) ([]core.Post, error) {
		if categorySlug != "news" {
			t.Fatalf("expected category 'news', got %q", categorySlug)
		}
		if limit != 50 { // clamp to maxPageSize
			t.Fatalf("expected clamped limit 50, got %d", limit)
		}
		return []core.Post{{Slug: "filtered"}}, nil
	}
	repo.listPublishedPostsSortedFn = func(context.Context, string, int32, int32) ([]core.Post, error) {
		t.Fatalf("should not call default list when category provided")
		return nil, nil
	}

	svc := NewService(repo)
	posts, err := svc.ListPublished(context.Background(), core.ListPostsOptions{Category: "news", Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 || posts[0].Slug != "filtered" {
		t.Fatalf("unexpected posts: %+v", posts)
	}
}

func TestServiceListPublished_InvalidSortFallsBack(t *testing.T) {
	repo := &fakePostRepo{}
	repo.listPublishedPostsSortedFn = func(ctx context.Context, sort string, limit, offset int32) ([]core.Post, error) {
		if sort != "created_at_desc" {
			t.Fatalf("expected fallback sort 'created_at_desc', got %q", sort)
		}
		return nil, nil
	}

	svc := NewService(repo)
	if _, err := svc.ListPublished(context.Background(), core.ListPostsOptions{Sort: "unknown"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceGetBySlugAggregates(t *testing.T) {
	repo := &fakePostRepo{}
	repo.getPostBySlugFn = func(ctx context.Context, slug string) (core.Post, error) {
		return core.Post{Slug: slug, Title: "Title"}, nil
	}
	repo.listCategoriesByPostSlugFn = func(ctx context.Context, slug string) ([]core.Category, error) {
		return []core.Category{{Slug: "golang"}}, nil
	}
	repo.listTagsByPostSlugFn = func(ctx context.Context, slug string) ([]core.Tag, error) {
		return []core.Tag{{Slug: "arch"}}, nil
	}

	svc := NewService(repo)
	result, err := svc.GetBySlug(context.Background(), "welcome")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Post.Slug != "welcome" || len(result.Categories) != 1 || len(result.Tags) != 1 {
		t.Fatalf("unexpected aggregate: %+v", result)
	}
}

func TestServiceGetBySlugValidates(t *testing.T) {
	svc := NewService(&fakePostRepo{})
	if _, err := svc.GetBySlug(context.Background(), "   "); !errors.Is(err, errSlugRequired) {
		t.Fatalf("expected errSlugRequired, got %v", err)
	}
}

func TestServiceCreateNormalizesCover(t *testing.T) {
	repo := &fakePostRepo{}
	repo.createPostFn = func(ctx context.Context, input core.CreatePostInput) (core.Post, error) {
		if input.CoverURL != nil {
			t.Fatalf("expected normalized CoverURL to nil, got %v", *input.CoverURL)
		}
		if input.Title != "Title" || input.Slug != "slug" {
			t.Fatalf("unexpected input: %+v", input)
		}
		return core.Post{ID: 1, Slug: input.Slug, Title: input.Title}, nil
	}

	svc := NewService(repo)
	cover := "   "
	post, err := svc.Create(context.Background(), core.CreatePostInput{Title: "Title", Slug: "slug", CoverURL: &cover})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.ID != 1 {
		t.Fatalf("expected created post, got %+v", post)
	}
}

func TestServiceCreateValidates(t *testing.T) {
	svc := NewService(&fakePostRepo{})
	if _, err := svc.Create(context.Background(), core.CreatePostInput{Slug: "slug"}); !errors.Is(err, errTitleRequired) {
		t.Fatalf("expected errTitleRequired, got %v", err)
	}
	if _, err := svc.Create(context.Background(), core.CreatePostInput{Title: "Title"}); !errors.Is(err, errSlugRequired) {
		t.Fatalf("expected errSlugRequired, got %v", err)
	}
}

func TestServiceUpdateNormalizes(t *testing.T) {
	repo := &fakePostRepo{}
	repo.updatePostBySlugFn = func(ctx context.Context, input core.UpdatePostInput) (core.Post, error) {
		if input.CoverURL != nil {
			t.Fatalf("expected CoverURL nil after normalization")
		}
		if input.Slug != "slug" || input.Title != "Updated" {
			t.Fatalf("unexpected update input: %+v", input)
		}
		return core.Post{Slug: input.Slug, Title: input.Title}, nil
	}

	svc := NewService(repo)
	cover := ""
	post, err := svc.Update(context.Background(), core.UpdatePostInput{Slug: " slug ", Title: " Updated ", CoverURL: &cover})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Title != "Updated" {
		t.Fatalf("expected updated title, got %+v", post)
	}
}

func TestServiceUpdateValidates(t *testing.T) {
	svc := NewService(&fakePostRepo{})
	if _, err := svc.Update(context.Background(), core.UpdatePostInput{Slug: ""}); !errors.Is(err, errSlugRequired) {
		t.Fatalf("expected errSlugRequired, got %v", err)
	}
	if _, err := svc.Update(context.Background(), core.UpdatePostInput{Slug: "slug"}); !errors.Is(err, errTitleRequired) {
		t.Fatalf("expected errTitleRequired, got %v", err)
	}
}

func TestServiceDeleteValidates(t *testing.T) {
	svc := NewService(&fakePostRepo{})
	if err := svc.Delete(context.Background(), " "); !errors.Is(err, errSlugRequired) {
		t.Fatalf("expected errSlugRequired, got %v", err)
	}
}

func TestServiceDeleteCallsRepo(t *testing.T) {
	called := false
	repo := &fakePostRepo{}
	repo.deletePostBySlugFn = func(ctx context.Context, slug string) error {
		called = true
		if slug != "slug" {
			t.Fatalf("unexpected slug: %q", slug)
		}
		return nil
	}

	svc := NewService(repo)
	if err := svc.Delete(context.Background(), "slug"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected delete to call repository")
	}
}

func TestServiceAddRemoveCategory(t *testing.T) {
	addCalled := false
	removeCalled := false
	repo := &fakePostRepo{}
	repo.addCategoryToPostFn = func(ctx context.Context, slug, categorySlug string) error {
		addCalled = true
		if slug != "slug" || categorySlug != "cat" {
			t.Fatalf("unexpected args: %s %s", slug, categorySlug)
		}
		return nil
	}
	repo.removeCategoryFromPostFn = func(ctx context.Context, slug, categorySlug string) error {
		removeCalled = true
		if slug != "slug" || categorySlug != "cat" {
			t.Fatalf("unexpected args: %s %s", slug, categorySlug)
		}
		return nil
	}

	svc := NewService(repo)
	if err := svc.AddCategory(context.Background(), "slug", "cat"); err != nil {
		t.Fatalf("AddCategory error: %v", err)
	}
	if err := svc.RemoveCategory(context.Background(), "slug", "cat"); err != nil {
		t.Fatalf("RemoveCategory error: %v", err)
	}
	if !addCalled || !removeCalled {
		t.Fatal("expected category methods to be called")
	}
}

func TestServiceAddRemoveTag(t *testing.T) {
	addCalled := false
	removeCalled := false
	repo := &fakePostRepo{}
	repo.addTagToPostFn = func(ctx context.Context, slug, tagSlug string) error {
		addCalled = true
		if slug != "slug" || tagSlug != "tag" {
			t.Fatalf("unexpected args: %s %s", slug, tagSlug)
		}
		return nil
	}
	repo.removeTagFromPostFn = func(ctx context.Context, slug, tagSlug string) error {
		removeCalled = true
		if slug != "slug" || tagSlug != "tag" {
			t.Fatalf("unexpected args: %s %s", slug, tagSlug)
		}
		return nil
	}

	svc := NewService(repo)
	if err := svc.AddTag(context.Background(), "slug", "tag"); err != nil {
		t.Fatalf("AddTag error: %v", err)
	}
	if err := svc.RemoveTag(context.Background(), "slug", "tag"); err != nil {
		t.Fatalf("RemoveTag error: %v", err)
	}
	if !addCalled || !removeCalled {
		t.Fatal("expected tag methods to be called")
	}
}

func TestServiceAddCategoryValidates(t *testing.T) {
	svc := NewService(&fakePostRepo{})
	if err := svc.AddCategory(context.Background(), "", "cat"); !errors.Is(err, errSlugRequired) {
		t.Fatalf("expected errSlugRequired, got %v", err)
	}
	if err := svc.AddCategory(context.Background(), "slug", " "); err == nil {
		t.Fatal("expected error for empty category slug")
	}
}

func TestServiceAddTagValidates(t *testing.T) {
	svc := NewService(&fakePostRepo{})
	if err := svc.AddTag(context.Background(), "", "tag"); !errors.Is(err, errSlugRequired) {
		t.Fatalf("expected errSlugRequired, got %v", err)
	}
	if err := svc.AddTag(context.Background(), "slug", " "); err == nil {
		t.Fatal("expected error for empty tag slug")
	}
}

type fakePostRepo struct {
	listPublishedPostsFn                 func(ctx context.Context, limit, offset int32) ([]core.Post, error)
	listPublishedPostsSortedFn           func(ctx context.Context, sort string, limit, offset int32) ([]core.Post, error)
	listPublishedPostsByCategorySortedFn func(ctx context.Context, categorySlug, sort string, limit, offset int32) ([]core.Post, error)
	listPublishedPostsByTagSortedFn      func(ctx context.Context, tagSlug, sort string, limit, offset int32) ([]core.Post, error)
	getPostBySlugFn                      func(ctx context.Context, slug string) (core.Post, error)
	createPostFn                         func(ctx context.Context, input core.CreatePostInput) (core.Post, error)
	updatePostBySlugFn                   func(ctx context.Context, input core.UpdatePostInput) (core.Post, error)
	deletePostBySlugFn                   func(ctx context.Context, slug string) error
	addCategoryToPostFn                  func(ctx context.Context, slug, categorySlug string) error
	removeCategoryFromPostFn             func(ctx context.Context, slug, categorySlug string) error
	addTagToPostFn                       func(ctx context.Context, slug, tagSlug string) error
	removeTagFromPostFn                  func(ctx context.Context, slug, tagSlug string) error
	listCategoriesByPostSlugFn           func(ctx context.Context, slug string) ([]core.Category, error)
	listTagsByPostSlugFn                 func(ctx context.Context, slug string) ([]core.Tag, error)
}

func (f *fakePostRepo) ListPublishedPosts(ctx context.Context, limit, offset int32) ([]core.Post, error) {
	if f.listPublishedPostsFn != nil {
		return f.listPublishedPostsFn(ctx, limit, offset)
	}
	return nil, nil
}

func (f *fakePostRepo) ListPublishedPostsSorted(ctx context.Context, sort string, limit, offset int32) ([]core.Post, error) {
	if f.listPublishedPostsSortedFn != nil {
		return f.listPublishedPostsSortedFn(ctx, sort, limit, offset)
	}
	return nil, nil
}

func (f *fakePostRepo) ListPublishedPostsByCategorySorted(ctx context.Context, categorySlug, sort string, limit, offset int32) ([]core.Post, error) {
	if f.listPublishedPostsByCategorySortedFn != nil {
		return f.listPublishedPostsByCategorySortedFn(ctx, categorySlug, sort, limit, offset)
	}
	return nil, nil
}

func (f *fakePostRepo) ListPublishedPostsByTagSorted(ctx context.Context, tagSlug, sort string, limit, offset int32) ([]core.Post, error) {
	if f.listPublishedPostsByTagSortedFn != nil {
		return f.listPublishedPostsByTagSortedFn(ctx, tagSlug, sort, limit, offset)
	}
	return nil, nil
}

func (f *fakePostRepo) GetPostBySlug(ctx context.Context, slug string) (core.Post, error) {
	if f.getPostBySlugFn != nil {
		return f.getPostBySlugFn(ctx, slug)
	}
	return core.Post{}, nil
}

func (f *fakePostRepo) CreatePost(ctx context.Context, input core.CreatePostInput) (core.Post, error) {
	if f.createPostFn != nil {
		return f.createPostFn(ctx, input)
	}
	return core.Post{}, nil
}

func (f *fakePostRepo) UpdatePostBySlug(ctx context.Context, input core.UpdatePostInput) (core.Post, error) {
	if f.updatePostBySlugFn != nil {
		return f.updatePostBySlugFn(ctx, input)
	}
	return core.Post{}, nil
}

func (f *fakePostRepo) DeletePostBySlug(ctx context.Context, slug string) error {
	if f.deletePostBySlugFn != nil {
		return f.deletePostBySlugFn(ctx, slug)
	}
	return nil
}

func (f *fakePostRepo) AddCategoryToPost(ctx context.Context, slug, categorySlug string) error {
	if f.addCategoryToPostFn != nil {
		return f.addCategoryToPostFn(ctx, slug, categorySlug)
	}
	return nil
}

func (f *fakePostRepo) RemoveCategoryFromPost(ctx context.Context, slug, categorySlug string) error {
	if f.removeCategoryFromPostFn != nil {
		return f.removeCategoryFromPostFn(ctx, slug, categorySlug)
	}
	return nil
}

func (f *fakePostRepo) AddTagToPost(ctx context.Context, slug, tagSlug string) error {
	if f.addTagToPostFn != nil {
		return f.addTagToPostFn(ctx, slug, tagSlug)
	}
	return nil
}

func (f *fakePostRepo) RemoveTagFromPost(ctx context.Context, slug, tagSlug string) error {
	if f.removeTagFromPostFn != nil {
		return f.removeTagFromPostFn(ctx, slug, tagSlug)
	}
	return nil
}

func (f *fakePostRepo) ListCategoriesByPostSlug(ctx context.Context, slug string) ([]core.Category, error) {
	if f.listCategoriesByPostSlugFn != nil {
		return f.listCategoriesByPostSlugFn(ctx, slug)
	}
	return nil, nil
}

func (f *fakePostRepo) ListTagsByPostSlug(ctx context.Context, slug string) ([]core.Tag, error) {
	if f.listTagsByPostSlugFn != nil {
		return f.listTagsByPostSlugFn(ctx, slug)
	}
	return nil, nil
}
