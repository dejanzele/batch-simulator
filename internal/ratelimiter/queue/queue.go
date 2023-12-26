package queue

type Queue[T any] interface {
	// Push adds the provided element to the queue.
	Push(item T) error
	// Pop removes and returns the first element in the queue.
	Pop() (T, error)
	// Len returns the number of elements in the queue.
	Len() int32
}

type WorkQueue[T any] interface {
	// Enqueue adds the provided elements to the queue.
	Enqueue(elems ...T) error
	// Dequeue removes and returns the first count elements in the queue.
	Dequeue(count int32) ([]T, error)
	// Len returns the number of elements in the queue.
	Len() int32
}
