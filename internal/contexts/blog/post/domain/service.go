package postdomain

import (
	"time"

	taxdomain "proto-gin-web/internal/contexts/blog/taxonomy/domain"
)

// ListPostsOptions describes pagination and filtering instructions.
type ListPostsOptions struct {
	Category string
	Tag      string
	Sort     string
	Limit    int32
	Offset   int32
}

// PostWithRelations bundles a post along with its taxonomy associations.
type PostWithRelations struct {
	Post       Post                  `json:"post"`
	Categories []taxdomain.Category  `json:"categories"`
	Tags       []taxdomain.Tag       `json:"tags"`
}

// CreatePostInput describes the data required to create a post.
type CreatePostInput struct {
	Title       string
	Slug        string
	Summary     string
	ContentMD   string
	CoverURL    *string
	Status      string
	AuthorID    int64
	PublishedAt *time.Time
}

// UpdatePostInput captures editable fields for an existing post identified by slug.
type UpdatePostInput struct {
	Slug      string
	Title     string
	Summary   string
	ContentMD string
	CoverURL  *string
	Status    string
}

