package presenter

import (
	"time"

	postdomain "proto-gin-web/internal/contexts/blog/post/domain"
	taxdomain "proto-gin-web/internal/contexts/blog/taxonomy/domain"
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
	cats := mapCategories(row.Categories)
	tags := mapTags(row.Tags)
	return PublicPostWithRelations{
		Post:       BuildPublicPost(row.Post),
		Categories: cats,
		Tags:       tags,
	}
}

func mapCategories(categories []taxdomain.Category) []PublicTaxonomy {
	result := make([]PublicTaxonomy, len(categories))
	for i, c := range categories {
		result[i] = PublicTaxonomy{ID: c.ID, Name: c.Name, Slug: c.Slug}
	}
	return result
}

func mapTags(tags []taxdomain.Tag) []PublicTaxonomy {
	result := make([]PublicTaxonomy, len(tags))
	for i, t := range tags {
		result[i] = PublicTaxonomy{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}
	return result
}

