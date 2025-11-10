package contenthttp

import "proto-gin-web/internal/domain"

// AdminPostResponse documents the admin post JSON envelope.
type AdminPostResponse struct {
	Ok   bool        `json:"ok"`
	Data domain.Post `json:"data"`
}

// AdminCategoryResponse documents the admin category JSON envelope.
type AdminCategoryResponse struct {
	Ok   bool            `json:"ok"`
	Data domain.Category `json:"data"`
}

// AdminTagResponse documents the admin tag JSON envelope.
type AdminTagResponse struct {
	Ok   bool       `json:"ok"`
	Data domain.Tag `json:"data"`
}

// AdminErrorResponse documents admin error messaging.
type AdminErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
