package api

import "proto-gin-web/internal/interfaces/http/view/presenter"

// postListResponse documents the JSON envelope returned by /api/posts.
type postListResponse struct {
	Ok   bool                   `json:"ok"`
	Data []presenter.PublicPost `json:"data"`
}

// postResponse documents the JSON envelope returned by /api/posts/{slug}.
type postResponse struct {
	Ok   bool                              `json:"ok"`
	Data presenter.PublicPostWithRelations `json:"data"`
}

// errorResponse documents the JSON error envelope.
type errorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
