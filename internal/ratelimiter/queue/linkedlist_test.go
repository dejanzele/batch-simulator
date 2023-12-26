package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLinkedListQueue(t *testing.T) {
	t.Parallel()

	t.Run("new empty queue", func(t *testing.T) {
		t.Parallel()

		q := NewLinkedListQueue[int]()
		assert.Nil(t, q.head)
		assert.Nil(t, q.tail)
	})

	t.Run("new queue with 1 element", func(t *testing.T) {
		t.Parallel()

		q := NewLinkedListQueue[int](1)
		assert.Equal(t, 1, q.head.val)
		assert.Equal(t, 1, q.tail.val)
		assert.Equal(t, q.head, q.tail)
	})

	t.Run("new queue with more than 1 element", func(t *testing.T) {
		t.Parallel()

		q := NewLinkedListQueue[int](1, 2, 3)
		assert.Equal(t, 1, q.head.val)
		assert.Equal(t, 3, q.tail.val)
		assert.Equal(t, q.head.next.next, q.tail)
	})
}

func TestLinkedList_Push(t *testing.T) {
	t.Parallel()

	t.Run("push in empty list", func(t *testing.T) {
		t.Parallel()

		q := &LinkedListQueue[int]{}
		if err := q.Push(1); err != nil {
			t.Fatalf("unexpected error when on Push: %v", err)
		}
		assert.Equal(t, 1, q.head.val)
		assert.Equal(t, 1, q.tail.val)
		assert.Equal(t, q.head, q.tail)
	})

	t.Run("push 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &LinkedListQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		assert.Equal(t, 1, q.head.val)
		assert.Equal(t, 2, q.tail.val)
		assert.Equal(t, q.head.next, q.tail)
	})

	t.Run("push more than 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &LinkedListQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		assert.NoError(t, q.Push(3))
		assert.NoError(t, q.Push(4))
		assert.Equal(t, 1, q.head.val)
		assert.Equal(t, 4, q.tail.val)
		assert.Equal(t, q.head.next.next.next, q.tail)
	})
}

func TestLinkedList_Pop(t *testing.T) {
	t.Parallel()

	t.Run("pop in empty list", func(t *testing.T) {
		t.Parallel()

		q := &LinkedListQueue[int]{}
		val, err := q.Pop()
		assert.NoError(t, err)
		assert.Equal(t, 0, val)
	})

	t.Run("pop from list with 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &LinkedListQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		val, err := q.Pop()
		assert.NoError(t, err)
		assert.Equal(t, 1, val)
		assert.Equal(t, 2, q.head.val)
		assert.Equal(t, 2, q.tail.val)
		assert.Equal(t, q.head, q.tail)
	})

	t.Run("pop from list with more than 2 elements", func(t *testing.T) {
		t.Parallel()

		q := &LinkedListQueue[int]{}
		assert.NoError(t, q.Push(1))
		assert.NoError(t, q.Push(2))
		assert.NoError(t, q.Push(3))
		assert.NoError(t, q.Push(4))

		val, err := q.Pop()
		if err != nil {
			t.Fatalf("unexpected error when on Pop: %v", err)
		}
		assert.Equal(t, 1, val)
		assert.Equal(t, 2, q.head.val)
		assert.Equal(t, 4, q.tail.val)
	})
}
