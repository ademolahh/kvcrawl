package main

import "sync"

type Queue struct {
	m    map[string]chan string
	mu   sync.Mutex
	size int
}

func NewQueue(size int) *Queue {
	return &Queue{
		m:    make(map[string]chan string),
		size: size,
	}
}

func (q *Queue) get(u string) (chan string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// filter for host

	ch, ok := q.m[u]
	if !ok {
		ch = make(chan string, q.size)
		q.m[u] = ch
	}

	return ch, ok
}

// fetch host here
