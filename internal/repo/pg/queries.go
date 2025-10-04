package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queries exposes typed helpers around the application database interactions.
type Queries struct {
	pool *pgxpool.Pool
}

// New wraps a pgx pool to provide strongly typed query helpers.
func New(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

type Article struct {
	ID          int64
	Title       string
	Body        string
	PublishedAt *time.Time
	AuthorID    int64
	CreatedAt   time.Time
}

type Post struct {
	ID          int64
	Title       string
	Slug        string
	Summary     string
	ContentMd   string
	CoverUrl    string
	Status      string
	AuthorID    int64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID   int64
	Name string
	Slug string
}

type Tag struct {
	ID   int64
	Name string
	Slug string
}

type CreatePostParams struct {
	Title       string
	Slug        string
	Summary     string
	ContentMd   string
	CoverUrl    *string
	Status      string
	AuthorID    int64
	PublishedAt *time.Time
}

type UpdatePostBySlugParams struct {
	Slug      string
	Title     string
	Summary   string
	ContentMd string
	CoverUrl  *string
	Status    string
}

type CreateCategoryParams struct {
	Name string
	Slug string
}

type CreateTagParams struct {
	Name string
	Slug string
}

func (q *Queries) ListArticles(ctx context.Context, limit, offset int32) ([]Article, error) {
	const stmt = `SELECT id, title, body, published_at, author_id, created_at FROM article ORDER BY id DESC LIMIT $1 OFFSET $2`
	rows, err := q.pool.Query(ctx, stmt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Article
	for rows.Next() {
		var a Article
		var published pgtype.Timestamptz
		if err := rows.Scan(&a.ID, &a.Title, &a.Body, &published, &a.AuthorID, &a.CreatedAt); err != nil {
			return nil, err
		}
		if published.Valid {
			t := published.Time
			a.PublishedAt = &t
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (q *Queries) ListPublishedPosts(ctx context.Context, limit, offset int32) ([]Post, error) {
	const stmt = `SELECT id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at FROM post WHERE status = 'published' ORDER BY COALESCE(published_at, created_at) DESC LIMIT $1 OFFSET $2`
	return q.listPosts(ctx, stmt, limit, offset)
}

func (q *Queries) ListPublishedPostsSorted(ctx context.Context, sort string, limit, offset int32) ([]Post, error) {
	const stmt = `SELECT id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at FROM post WHERE status = 'published' ORDER BY CASE WHEN $1 = 'published_at_asc' THEN published_at END ASC, CASE WHEN $1 = 'published_at_desc' THEN published_at END DESC, CASE WHEN $1 = 'created_at_asc' THEN created_at END ASC, CASE WHEN $1 = 'created_at_desc' OR $1 = '' THEN created_at END DESC NULLS LAST LIMIT $2 OFFSET $3`
	return q.listPosts(ctx, stmt, sort, limit, offset)
}

func (q *Queries) ListPublishedPostsByCategorySorted(ctx context.Context, slug string, sort string, limit, offset int32) ([]Post, error) {
	const stmt = `SELECT p.id, p.title, p.slug, p.summary, p.content_md, p.cover_url, p.status, p.author_id, p.published_at, p.created_at, p.updated_at FROM post p JOIN post_category pc ON pc.post_id = p.id JOIN category c ON c.id = pc.category_id WHERE p.status = 'published' AND c.slug = $1 ORDER BY CASE WHEN $2 = 'published_at_asc' THEN p.published_at END ASC, CASE WHEN $2 = 'published_at_desc' THEN p.published_at END DESC, CASE WHEN $2 = 'created_at_asc' THEN p.created_at END ASC, CASE WHEN $2 = 'created_at_desc' OR $2 = '' THEN p.created_at END DESC NULLS LAST LIMIT $3 OFFSET $4`
	return q.listPosts(ctx, stmt, slug, sort, limit, offset)
}

func (q *Queries) ListPublishedPostsByTagSorted(ctx context.Context, slug string, sort string, limit, offset int32) ([]Post, error) {
	const stmt = `SELECT p.id, p.title, p.slug, p.summary, p.content_md, p.cover_url, p.status, p.author_id, p.published_at, p.created_at, p.updated_at FROM post p JOIN post_tag pt ON pt.post_id = p.id JOIN tag t ON t.id = pt.tag_id WHERE p.status = 'published' AND t.slug = $1 ORDER BY CASE WHEN $2 = 'published_at_asc' THEN p.published_at END ASC, CASE WHEN $2 = 'published_at_desc' THEN p.published_at END DESC, CASE WHEN $2 = 'created_at_asc' THEN p.created_at END ASC, CASE WHEN $2 = 'created_at_desc' OR $2 = '' THEN p.created_at END DESC NULLS LAST LIMIT $3 OFFSET $4`
	return q.listPosts(ctx, stmt, slug, sort, limit, offset)
}

func (q *Queries) GetPostBySlug(ctx context.Context, slug string) (Post, error) {
	const stmt = `SELECT id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at FROM post WHERE slug = $1`
	row := q.pool.QueryRow(ctx, stmt, slug)
	return scanPost(row)
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	const stmt = `INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at`
	var cover any
	if arg.CoverUrl != nil {
		cover = *arg.CoverUrl
	}
	var published any
	if arg.PublishedAt != nil {
		published = *arg.PublishedAt
	}
	row := q.pool.QueryRow(ctx, stmt, arg.Title, arg.Slug, arg.Summary, arg.ContentMd, cover, arg.Status, arg.AuthorID, published)
	return scanPost(row)
}

func (q *Queries) UpdatePostBySlug(ctx context.Context, arg UpdatePostBySlugParams) (Post, error) {
	const stmt = `UPDATE post SET title = $2, summary = $3, content_md = $4, cover_url = $5, status = $6, updated_at = NOW() WHERE slug = $1 RETURNING id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at`
	var cover any
	if arg.CoverUrl != nil {
		cover = *arg.CoverUrl
	}
	row := q.pool.QueryRow(ctx, stmt, arg.Slug, arg.Title, arg.Summary, arg.ContentMd, cover, arg.Status)
	return scanPost(row)
}

func (q *Queries) DeletePostBySlug(ctx context.Context, slug string) error {
	const stmt = `DELETE FROM post WHERE slug = $1`
	_, err := q.pool.Exec(ctx, stmt, slug)
	return err
}

func (q *Queries) AddCategoryToPost(ctx context.Context, slug, categorySlug string) error {
	const stmt = `INSERT INTO post_category (post_id, category_id) SELECT p.id, c.id FROM post p, category c WHERE p.slug = $1 AND c.slug = $2 ON CONFLICT DO NOTHING`
	_, err := q.pool.Exec(ctx, stmt, slug, categorySlug)
	return err
}

func (q *Queries) RemoveCategoryFromPost(ctx context.Context, slug, categorySlug string) error {
	const stmt = `DELETE FROM post_category USING post p, category c WHERE post_category.post_id = p.id AND post_category.category_id = c.id AND p.slug = $1 AND c.slug = $2`
	_, err := q.pool.Exec(ctx, stmt, slug, categorySlug)
	return err
}

func (q *Queries) AddTagToPost(ctx context.Context, slug, tagSlug string) error {
	const stmt = `INSERT INTO post_tag (post_id, tag_id) SELECT p.id, t.id FROM post p, tag t WHERE p.slug = $1 AND t.slug = $2 ON CONFLICT DO NOTHING`
	_, err := q.pool.Exec(ctx, stmt, slug, tagSlug)
	return err
}

func (q *Queries) RemoveTagFromPost(ctx context.Context, slug, tagSlug string) error {
	const stmt = `DELETE FROM post_tag USING post p, tag t WHERE post_tag.post_id = p.id AND post_tag.tag_id = t.id AND p.slug = $1 AND t.slug = $2`
	_, err := q.pool.Exec(ctx, stmt, slug, tagSlug)
	return err
}

func (q *Queries) ListCategoriesByPostSlug(ctx context.Context, slug string) ([]Category, error) {
	const stmt = `SELECT c.id, c.name, c.slug FROM category c JOIN post_category pc ON pc.category_id = c.id JOIN post p ON p.id = pc.post_id WHERE p.slug = $1 ORDER BY c.name ASC`
	rows, err := q.pool.Query(ctx, stmt, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (q *Queries) ListTagsByPostSlug(ctx context.Context, slug string) ([]Tag, error) {
	const stmt = `SELECT t.id, t.name, t.slug FROM tag t JOIN post_tag pt ON pt.tag_id = t.id JOIN post p ON p.id = pt.post_id WHERE p.slug = $1 ORDER BY t.name ASC`
	rows, err := q.pool.Query(ctx, stmt, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (q *Queries) CreateCategory(ctx context.Context, arg CreateCategoryParams) (Category, error) {
	const stmt = `INSERT INTO category (name, slug) VALUES ($1, $2) RETURNING id, name, slug`
	var c Category
	err := q.pool.QueryRow(ctx, stmt, arg.Name, arg.Slug).Scan(&c.ID, &c.Name, &c.Slug)
	return c, err
}

func (q *Queries) DeleteCategoryBySlug(ctx context.Context, slug string) error {
	const stmt = `DELETE FROM category WHERE slug = $1`
	_, err := q.pool.Exec(ctx, stmt, slug)
	return err
}

func (q *Queries) CreateTag(ctx context.Context, arg CreateTagParams) (Tag, error) {
	const stmt = `INSERT INTO tag (name, slug) VALUES ($1, $2) RETURNING id, name, slug`
	var t Tag
	err := q.pool.QueryRow(ctx, stmt, arg.Name, arg.Slug).Scan(&t.ID, &t.Name, &t.Slug)
	return t, err
}

func (q *Queries) DeleteTagBySlug(ctx context.Context, slug string) error {
	const stmt = `DELETE FROM tag WHERE slug = $1`
	_, err := q.pool.Exec(ctx, stmt, slug)
	return err
}

func (q *Queries) listPosts(ctx context.Context, stmt string, args ...any) ([]Post, error) {
	rows, err := q.pool.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func scanPost(row pgx.Row) (Post, error) {
	var p Post
	var cover sql.NullString
	var published pgtype.Timestamptz
	if err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.ContentMd, &cover, &p.Status, &p.AuthorID, &published, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return Post{}, err
	}
	if cover.Valid {
		p.CoverUrl = cover.String
	}
	if published.Valid {
		t := published.Time
		p.PublishedAt = &t
	}
	return p, nil
}
