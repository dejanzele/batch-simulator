package ratelimiter

import "context"

// Executor defines how a work item should be processed.
type Executor[T any] interface {
	// Identifier returns the identifier of the executor.
	Identifier() string
	// Execute defines the logic how the work item should be executed.
	Execute(ctx context.Context, item T) error
}
