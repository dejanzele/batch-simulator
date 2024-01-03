package queue

type ArrayQueue[T any] struct {
	queue   []T
	current int32
}

// NewArrayQueue returns a new queue with the provided elements.
func NewArrayQueue[T any](elems []T) *ArrayQueue[T] {
	current := int32(0)
	if len(elems) > 0 {
		current = int32(len(elems))
	}
	return &ArrayQueue[T]{queue: elems, current: current}
}

func (a *ArrayQueue[T]) Push(item T) error {
	a.queue = append(a.queue, item)
	a.current++
	return nil
}

func (a *ArrayQueue[T]) Pop() (T, error) {
	if a.current == 0 {
		var zeroValue T
		return zeroValue, nil
	}
	a.current--
	return a.queue[a.current], nil
}

func (a *ArrayQueue[T]) Len() int32 {
	return a.current
}

func (a *ArrayQueue[T]) Enqueue(elems []T) error {
	copy(a.queue[a.current:], elems)
	return nil
}

func (a *ArrayQueue[T]) Dequeue(count int32) ([]T, error) {
	available := min(count, a.Len())
	elems := a.queue[a.current-available : a.current]
	a.current -= available
	return elems, nil
}

var _ Queue[any] = &ArrayQueue[any]{}
var _ WorkQueue[any] = &ArrayQueue[any]{}
