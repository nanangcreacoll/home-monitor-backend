package pkg

import "sync"

type Queue interface {
	Push(item any)
	Pop() any
	Size() int
	Empty() bool
}

type queue struct {
	mutex sync.Mutex
	items []any
}

func NewQueue() Queue {
	return &queue{items: make([]any, 0)}
}

func (q *queue) Push(item any) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
}

func (q *queue) Pop() any {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *queue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}

func (q *queue) Empty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items) == 0
}
