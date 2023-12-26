package ratelimiter

import (
	"context"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestRateLimiter_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("stops if started param updated to false", func(t *testing.T) {
		t.Parallel()

		rl := New[int](1*time.Second, 1, newNoopExecutor[int]())

		go rl.Run(ctx)

		time.Sleep(10 * time.Millisecond)

		assert.True(t, rl.started)

		rl.started = false

		time.Sleep(10 * time.Millisecond)

		assert.False(t, rl.started)
	})

	t.Run("stops if context is cancelled", func(t *testing.T) {
		t.Parallel()

		rl := New[int](2*time.Millisecond, 1, newNoopExecutor[int]())
		ctx, cancel := context.WithCancel(ctx)
		go rl.Run(ctx)

		time.Sleep(10 * time.Millisecond)

		assert.True(t, rl.started)

		cancel()

		time.Sleep(10 * time.Millisecond)

		assert.False(t, rl.started)
		assert.Equal(t, context.Canceled, ctx.Err())
		assert.Equal(t, rl.metrics.Executed, int32(0))
		assert.Equal(t, rl.metrics.Succeeded, int32(0))
		assert.Equal(t, rl.metrics.Failed, int32(0))
	})

	t.Run("executes work items", func(t *testing.T) {
		t.Parallel()

		ex := newCacheExecutor()
		rl := New[int](20*time.Millisecond, 3, ex)
		assert.NoError(t, rl.queue.Enqueue(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))

		go rl.Run(ctx)

		assert.Eventually(
			t,
			func() bool {
				return len(ex.Cache) == 10
			},
			400*time.Millisecond,
			20*time.Millisecond,
		)
		assert.ElementsMatch(t, ex.Cache, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		assert.Equal(t, rl.metrics.Executed, int32(10))
		assert.Equal(t, rl.metrics.Succeeded, int32(10))
		assert.Equal(t, rl.metrics.Failed, int32(0))
	})

	t.Run("executes work items with timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(ctx, 90*time.Millisecond)
		defer cancel()

		ex := newCacheExecutor()
		rl := New[int](20*time.Millisecond, 1, ex)
		assert.NoError(t, rl.queue.Enqueue(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))

		go rl.Run(ctx)

		assert.Eventually(
			t,
			func() bool {
				return len(ex.Cache) == 4
			},
			200*time.Millisecond,
			20*time.Millisecond,
		)
		assert.ElementsMatch(t, ex.Cache, []int{1, 2, 3, 4})
		assert.Equal(t, rl.metrics.Executed, int32(4))
		assert.Equal(t, rl.metrics.Succeeded, int32(4))
		assert.Equal(t, rl.metrics.Failed, int32(0))
	})

	t.Run("executor returns error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(ctx, 15*time.Millisecond)
		defer cancel()

		ex := newErrorExecutor()
		rl := New[int](10*time.Millisecond, 1, ex)
		assert.NoError(t, rl.queue.Enqueue(1))

		go rl.Run(ctx)

		select {
		case err := <-rl.ErrChan():
			var createError *CreateError
			assert.ErrorAs(t, err, &createError)
			assert.Equal(t, "v1", createError.APIGroup)
			assert.Equal(t, "Pod", createError.Kind)
			assert.Equal(t, "test-pod", createError.Resource.GetName())
			assert.ErrorIs(t, createError.Err, assert.AnError)

		case <-ctx.Done():
			t.Fatal("failed to receive error from errChan in given time")
		}
		assert.Equal(t, rl.metrics.Executed, int32(1))
		assert.Equal(t, rl.metrics.Succeeded, int32(0))
		assert.Equal(t, rl.metrics.Failed, int32(1))
	})

}

type noopExecutor[T any] struct{}

func newNoopExecutor[T any]() *noopExecutor[T] {
	return &noopExecutor[T]{}
}

func (n *noopExecutor[T]) Identifier() string {
	return "noop"
}

func (n *noopExecutor[T]) Execute(ctx context.Context, item T) error {
	return nil
}

type cacheExecutor[T int] struct {
	Cache []int
}

func newCacheExecutor() *cacheExecutor[int] {
	return &cacheExecutor[int]{}
}

func (c *cacheExecutor[T]) Identifier() string {
	return "counter"
}

func (c *cacheExecutor[T]) Execute(ctx context.Context, item int) error {
	c.Cache = append(c.Cache, item)
	return nil
}

type errorExecutor[T int] struct{}

func newErrorExecutor() *errorExecutor[int] {
	return &errorExecutor[int]{}
}

func (e *errorExecutor[T]) Identifier() string {
	return "error"
}

func (e *errorExecutor[T]) Execute(ctx context.Context, item int) error {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}
	return NewCreateError(assert.AnError, "v1", "Pod", pod)
}
