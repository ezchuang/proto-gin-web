package presenter

import (
	"time"

	postdomain "proto-gin-web/internal/blog/post/domain"
)

// PublicPost represents the public JSON shape of a blog post.
type PublicPost struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Summary     string     `json:"summary"`
	ContentMD   string     `json:"content_md"`
	CoverURL    string     `json:"cover_url"`
	Status      string     `json:"status"`
	AuthorID    int64      `json:"author_id"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PublicTaxonomy describes the external JSON shape of category/tag.
type PublicTaxonomy struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// PublicPostWithRelations contains a post and its taxonomies.
type PublicPostWithRelations struct {
	Post       PublicPost       `json:"post"`
	Categories []PublicTaxonomy `json:"categories"`
	Tags       []PublicTaxonomy `json:"tags"`
}

// BuildPublicPosts converts domain posts into public representation.
func BuildPublicPosts(posts []postdomain.Post) []PublicPost {
	result := make([]PublicPost, len(posts))
	for i, p := range posts {
		result[i] = BuildPublicPost(p)
	}
	return result
}

// BuildPublicPost converts a single post.
func BuildPublicPost(p postdomain.Post) PublicPost {
	return PublicPost{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Summary:     p.Summary,
		ContentMD:   p.ContentMD,
		CoverURL:    p.CoverURL,
		Status:      p.Status,
		AuthorID:    p.AuthorID,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// BuildPublicPostWithRelations converts a domain result to public shape.
func BuildPublicPostWithRelations(row postdomain.PostWithRelations) PublicPostWithRelations {
	cats := make([]PublicTaxonomy, len(row.Categories))
	for i, cat := range row.Categories {
		cats[i] = PublicTaxonomy{ID: cat.ID, Name: cat.Name, Slug: cat.Slug}
	}
	tags := make([]PublicTaxonomy, len(row.Tags))
	for i, tag := range row.Tags {
		tags[i] = PublicTaxonomy{ID: tag.ID, Name: tag.Name, Slug: tag.Slug}
	}
	return PublicPostWithRelations{
		Post:       BuildPublicPost(row.Post),
		Categories: cats,
		Tags:       tags,
	}
}
