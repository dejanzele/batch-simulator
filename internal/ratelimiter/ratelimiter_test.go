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

		rl := New[int](1*time.Second, 1, 1, newNoopExecutor())

		go rl.Run(ctx)

		time.Sleep(10 * time.Millisecond)

		assert.True(t, rl.started)

		rl.started = false

		time.Sleep(10 * time.Millisecond)

		assert.False(t, rl.started)
	})

	t.Run("stops if context is cancelled", func(t *testing.T) {
		t.Parallel()

		rl := New[int](2*time.Millisecond, 1, 5, newNoopExecutor())
		ctx, cancel := context.WithCancel(ctx)
		go rl.Run(ctx)

		time.Sleep(10 * time.Millisecond)

		assert.True(t, rl.started)

		cancel()

		time.Sleep(10 * time.Millisecond)

		assert.False(t, rl.started)
		assert.Equal(t, context.Canceled, ctx.Err())
		assert.Equal(t, rl.metrics.Executed, 5)
		assert.Equal(t, rl.metrics.Succeeded, 5)
		assert.Equal(t, rl.metrics.Failed, 0)
	})

	t.Run("executes work items", func(t *testing.T) {
		t.Parallel()

		ex := newCacheExecutor()
		rl := New[int](20*time.Millisecond, 3, 10, ex)

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
		assert.Equal(t, rl.metrics.Executed, 10)
		assert.Equal(t, rl.metrics.Succeeded, 10)
		assert.Equal(t, rl.metrics.Failed, 0)
	})

	t.Run("executes work items with timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(ctx, 90*time.Millisecond)
		defer cancel()

		ex := newCacheExecutor()
		rl := New[int](20*time.Millisecond, 1, 5, ex)

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
		assert.Equal(t, rl.metrics.Executed, 4)
		assert.Equal(t, rl.metrics.Succeeded, 4)
		assert.Equal(t, rl.metrics.Failed, 0)
	})

	t.Run("executor returns error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(ctx, 175*time.Millisecond)
		defer cancel()

		ex := newErrorExecutor()
		rl := New[int](100*time.Millisecond, 1, 5, ex)

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
		assert.Equal(t, rl.metrics.Executed, 1)
		assert.Equal(t, rl.metrics.Succeeded, 0)
		assert.Equal(t, rl.metrics.Failed, 1)
	})

}

type noopExecutor struct{}

func newNoopExecutor() *noopExecutor {
	return &noopExecutor{}
}

func (n *noopExecutor) Identifier() string {
	return "noop"
}

func (n *noopExecutor) Execute(ctx context.Context) error {
	return nil
}

type cacheExecutor struct {
	Current int
	Cache   []int
}

func newCacheExecutor() *cacheExecutor {
	return &cacheExecutor{}
}

func (c *cacheExecutor) Identifier() string {
	return "counter"
}

func (c *cacheExecutor) Execute(ctx context.Context) error {
	c.Current++
	c.Cache = append(c.Cache, c.Current)
	return nil
}

type errorExecutor struct{}

func newErrorExecutor() *errorExecutor {
	return &errorExecutor{}
}

func (e *errorExecutor) Identifier() string {
	return "error"
}

func (e *errorExecutor) Execute(ctx context.Context) error {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}
	return NewCreateError(assert.AnError, "v1", "Pod", pod)
}
