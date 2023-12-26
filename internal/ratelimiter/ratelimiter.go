package ratelimiter

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dejanzele/batch-simulator/internal/ratelimiter/queue"
)

// RateLimiter is used to limit the rate at which work items are processed.
type RateLimiter[T any] struct {
	// started indicates whether the rate limiter is currently running.
	started bool
	// queue is the queue that should be used to store work items.
	queue queue.WorkQueue[T]
	// executor is the executor that should be used to process work items.
	executor Executor[T]
	// ticker is the ticker that should be used to trigger the rate limiter.
	ticker *time.Ticker
	// interval is the interval at which the rate limiter should be triggered.
	interval time.Duration
	// requests is the number of work items to process per interval.
	requests int32
	// logger is the logger that should be used to log messages.
	logger *slog.Logger
	// errChan is the channel that should be used to send errors.
	errChan chan error
	// metrics tracks the total number of work items that have been processed along with the number of work items that have failed and succeeded.
	metrics Metrics
	// mutex is used to synchronize access to the metrics.
	mutex sync.RWMutex
	// limit is the maximum number of work items that the rate limiter can process.
	// After the limit is reached, the rate limiter will stop processing work items.
	// If limit is 0, then there is no limit.
	limit int32
}

type Option[T any] func(*RateLimiter[T])

func WithLogger[T any](logger *slog.Logger) Option[T] {
	return func(r *RateLimiter[T]) {
		r.logger = logger
	}
}

func WithQueue[T any](q queue.WorkQueue[T]) Option[T] {
	return func(r *RateLimiter[T]) {
		r.queue = q
	}
}

func WithLimit[T any](limit int32) Option[T] {
	return func(r *RateLimiter[T]) {
		r.limit = limit
	}
}

// New creates a new RateLimiter.
// - frequency: the frequency of the rate limiter.
// - requests: the number of work items to process per interval.
// - executor: the executor to use to process the work items.
// - opts: the options to configure the RateLimiter.
func New[T any](frequency time.Duration, requests int32, executor Executor[T], opts ...Option[T]) *RateLimiter[T] {
	rl := &RateLimiter[T]{interval: frequency, requests: requests, executor: executor, errChan: make(chan error)}
	for _, opt := range opts {
		opt(rl)
	}
	if rl.logger == nil {
		rl.logger = &slog.Logger{}
	}
	if rl.queue == nil {
		rl.queue = queue.NewLinkedListQueue[T]()
	}
	rl.logger = slog.With("process", "ratelimiter", "executor", rl.executor.Identifier())
	return rl
}

// Run starts the rate limiter.
func (r *RateLimiter[T]) Run(ctx context.Context) {
	r.ticker = time.NewTicker(r.interval)
	defer r.ticker.Stop()

	r.logger.Info("starting ratelimiter")
	r.started = true
	for r.started {
		select {
		case <-ctx.Done():
			r.Stop()
			return
		case <-r.ticker.C:
			go func() {
				r.execute(ctx, r.errChan)
			}()
		}
	}
}

// Stop stops the rate limiter.
func (r *RateLimiter[T]) Stop() {
	r.logger.Info("stopping ratelimiter")
	r.started = false
}

// IsRunning returns true if the rate limiter is currently running.
func (r *RateLimiter[T]) IsRunning() bool {
	return r.started
}

// AddWorkItems adds work items to the queue.
func (r *RateLimiter[T]) AddWorkItems(items ...T) error {
	return r.queue.Enqueue(items...)
}

// ErrChan returns the channel that should be used to receive errors.
func (r *RateLimiter[T]) ErrChan() <-chan error {
	return r.errChan
}

// Metrics returns the metrics of the rate limiter.
func (r *RateLimiter[T]) Metrics() Metrics {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.metrics
}

// execute fetches work items from queue and sends them to executor for processing.
func (r *RateLimiter[T]) execute(ctx context.Context, errCh chan<- error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	executedSoFar := r.metrics.Executed
	isLimitReached := executedSoFar >= r.limit
	r.logger.Info("executing work items", "executed", executedSoFar, "limit", r.limit)
	if isLimitReached {
		r.logger.Info("maximum number of processed work items has been reached")
		r.Stop()
		return
	}
	items, _ := r.queue.Dequeue(r.requests)
	var executed, failed, succeeded int32
	r.logger.Info("processing work items", "dequeued", len(items), "requested", r.requests)
	for i, item := range items {
		r.logger.Debug("executing work item", "index", i)
		executed++
		if err := r.executor.Execute(ctx, item); err != nil {
			failed++
			errCh <- fmt.Errorf("failed to execute work item: %w", err)
		} else {
			succeeded++
		}
	}
	if len(items) > 0 {
		r.metrics.Add(executed, failed, succeeded)
	}
}
