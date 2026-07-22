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

	if _, ok := s.m[u]; ok {
		return false
	}

	s.m[u] = struct{}{}
	return true
}

func (s *Seen) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.m)
}

func (c *Crawler) maximumReached() bool {
	c.url.mu.Lock()
	defer c.url.mu.Unlock()
	return len(c.url.m) >= c.maximum
}
