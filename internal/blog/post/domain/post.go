package postdomain

import "time"

// Post is the blog domain entity.
type Post struct {
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
