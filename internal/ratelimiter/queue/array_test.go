package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewArrayQueue(t *testing.T) {
	t.Parallel()

	t.Run("new empty queue", func(t *testing.T) {
		t.Parallel()

		q := NewArrayQueue[int](nil)
		assert.Nil(t, q.queue)
		assert.Equal(t, int32(0), q.current)
	})

	t.Run("new queue with 1 element", func(t *testing.T) {
		t.Parallel()

		q := NewArrayQueue[int]([]int{1})
		assert.Len(t, q.queue, 1)
		assert.Equal(t, 1, q.queue[0])
		assert.Equal(t, int32(1), q.current)
	})

	t.Run("new queue with more than 1 element", func(t *testing.T) {
		t.Parallel()

		q := NewArrayQueue[int]([]int{1, 2, 3})
		assert.Len(t, q.queue, 3)
		assert.ElementsMatch(t, []int{1, 2, 3}, q.queue)
		assert.Equal(t, int32(3), q.current)
	})
}

func TestArray_Push(t *testing.T) {
	t.Parallel()

	t.Run("push in empty list", func(t *testing.T) {
		t.Parallel()

		q := &ArrayQueue[int]{}
		if err := q.Push(1); err != nil {
			t.Fatalf("unexpected error when on Push: %v", err)
		}
		assert.Len(t, q.queue, 1)
		assert.Equal(t, 1, q.queue[0])
		assert.Equal(t, int32(1), q.current)
	})

	t.Run("push 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &ArrayQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		assert.Len(t, q.queue, 2)
		assert.Equal(t, q.queue[0], 1)
		assert.Equal(t, q.queue[1], 2)
		assert.Equal(t, int32(2), q.current)
	})

	t.Run("push more than 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &ArrayQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		assert.NoError(t, q.Push(3))
		assert.NoError(t, q.Push(4))
		assert.Len(t, q.queue, 4)
		assert.ElementsMatch(t, []int{1, 2, 3, 4}, q.queue)
		assert.Equal(t, int32(4), q.current)
	})
}

func TestArray_Pop(t *testing.T) {
	t.Parallel()

	t.Run("pop in empty list", func(t *testing.T) {
		t.Parallel()

		q := &ArrayQueue[int]{}
		val, err := q.Pop()
		assert.NoError(t, err)
		assert.Equal(t, 0, val)
		assert.Equal(t, int32(0), q.current)
	})

	t.Run("pop from list with 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &ArrayQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		val, err := q.Pop()
		assert.NoError(t, err)
		assert.Equal(t, 2, val)
		assert.Equal(t, q.Len(), int32(1))
		assert.Equal(t, int32(1), q.current)
	})

	t.Run("pop from list with more than 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &ArrayQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		assert.NoError(t, q.Push(3))
		assert.NoError(t, q.Push(4))

		val, err := q.Pop()
		if err != nil {
			t.Fatalf("unexpected error when on Pop: %v", err)
		}
		assert.Equal(t, 4, val)
		assert.Equal(t, q.Len(), int32(3))
		assert.Equal(t, int32(3), q.current)
	})
}
