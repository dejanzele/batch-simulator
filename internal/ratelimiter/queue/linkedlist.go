package queue

type LinkedListQueue[T any] struct {
	head *node[T]
	tail *node[T]
	len  int32
}

// NewLinkedListQueue returns a new queue with the provided elements.
func NewLinkedListQueue[T any](elems ...T) *LinkedListQueue[T] {
	l := &LinkedListQueue[T]{}
	for _, e := range elems {
		_ = l.Push(e)
	}
	return l
}

type node[T any] struct {
	next *node[T]
	val  T
}

func newNode[T any](e T) *node[T] {
	return &node[T]{val: e}
}

// Push adds the provided element to the queue.
func (q *LinkedListQueue[T]) Push(e T) error {
	n := newNode(e)
	if q.head == nil {
		q.head = n
		q.tail = q.head
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.len++
	return nil
}

// Pop removes and returns the first element in the queue.
func (q *LinkedListQueue[T]) Pop() (T, error) {
	if q.head == nil {
		var zeroValue T
		return zeroValue, nil
	}
	ret := q.head.val
	q.head = q.head.next
	q.len--
	return ret, nil
}

// Len returns the number of elements in the queue.
func (q *LinkedListQueue[T]) Len() int32 {
	return q.len
}

// Enqueue adds the provided elements to the queue.
func (q *LinkedListQueue[T]) Enqueue(elems ...T) error {
	for _, e := range elems {
		if err := q.Push(e); err != nil {
			return err
		}
	}
	return nil
}

// Dequeue returns the first elements specified by the count param.
// If count is greater than the number of elements in the queue, all elements are returned.
func (q *LinkedListQueue[T]) Dequeue(count int32) ([]T, error) {
	available := min(count, q.Len())
	items := make([]T, available)
	for i := int32(0); i < available; i++ {
		items[i], _ = q.Pop()
	}
	return items, nil
}

var _ Queue[any] = &LinkedListQueue[any]{}
