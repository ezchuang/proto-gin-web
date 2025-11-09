package api

import "proto-gin-web/internal/domain"

// postListResponse documents the JSON envelope returned by /api/posts.
type postListResponse struct {
	Ok   bool          `json:"ok"`
	Data []domain.Post `json:"data"`
}

// postResponse documents the JSON envelope returned by /api/posts/{slug}.
type postResponse struct {
	Ok   bool                     `json:"ok"`
	Data domain.PostWithRelations `json:"data"`
}

// errorResponse documents the JSON error envelope.
type errorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
