package core

import "context"

// PostService defines blog use cases (thin, extend later).
type PostService interface {
	// List public posts with pagination
	List(ctx context.Context, limit, offset int32) ([]Post, error)
}
