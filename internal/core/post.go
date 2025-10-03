package core

import "time"

// Post is the enterprise blog core entity.
type Post struct {
	ID          int64
	Title       string
	Slug        string
	Summary     string
	ContentMD   string
	CoverURL    string
	Status      string // draft, published, archived
	AuthorID    int64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
