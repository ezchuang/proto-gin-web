package api

import presenters "proto-gin-web/internal/contexts/blog/post/adapters/view"

// postListResponse documents the JSON envelope returned by /api/posts.
type postListResponse struct {
	Ok   bool                    `json:"ok"`
	Data []presenters.PublicPost `json:"data"`
}

// postResponse documents the JSON envelope returned by /api/posts/{slug}.
type postResponse struct {
	Ok   bool                               `json:"ok"`
	Data presenters.PublicPostWithRelations `json:"data"`
}

// errorResponse documents the JSON error envelope.
type errorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

