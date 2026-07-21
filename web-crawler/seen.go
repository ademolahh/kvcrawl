package main

import "sync"

type Seen struct {
	m  map[string]struct{}
	mu sync.Mutex
}

func NewSeen() *Seen {
	return &Seen{
		m: make(map[string]struct{}),
	}
}

func (s *Seen) tryAdd(u string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[u]; !ok {
		return false
	}

	s.m[u] = struct{}{}
	return true
}
