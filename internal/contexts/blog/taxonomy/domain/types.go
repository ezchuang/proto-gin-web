package taxdomain

// Category represents a taxonomy group assigned to posts.
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Tag is a flexible label attached to posts.
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// CreateCategoryInput holds fields required to create a category.
type CreateCategoryInput struct {
	Name string
	Slug string
}

// CreateTagInput holds fields required to create a tag.
type CreateTagInput struct {
	Name string
	Slug string
}

