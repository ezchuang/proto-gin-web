package admincontent

import (
	postdomain "proto-gin-web/internal/contexts/blog/post/domain"
	taxdomain "proto-gin-web/internal/contexts/blog/taxonomy/domain"
)

// AdminPostResponse documents the admin post JSON envelope.
type AdminPostResponse struct {
	Ok   bool            `json:"ok"`
	Data postdomain.Post `json:"data"`
}

// AdminCategoryResponse documents the admin category JSON envelope.
type AdminCategoryResponse struct {
	Ok   bool               `json:"ok"`
	Data taxdomain.Category `json:"data"`
}

// AdminTagResponse documents the admin tag JSON envelope.
type AdminTagResponse struct {
	Ok   bool          `json:"ok"`
	Data taxdomain.Tag `json:"data"`
}

// AdminErrorResponse documents admin error messaging.
type AdminErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

