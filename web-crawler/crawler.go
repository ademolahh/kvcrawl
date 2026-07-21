package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sync"

	"golang.org/x/net/html"
)

type Crawler struct {
	content  *Seen
	url      *Seen
	queue    *Queue
	capacity int
	maximum  int
	tasks    sync.WaitGroup
	workers  sync.WaitGroup
}

func NewCrawler(capacity, maximum, queueSize int) *Crawler {
	return &Crawler{
		content:  NewSeen(),
		url:      NewSeen(),
		queue:    NewQueue(queueSize),
		capacity: capacity,
		maximum:  maximum,
	}
}

func (c *Crawler) download(u string) {
	if !c.url.tryAdd(u) {
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", u, nil)
	if err != nil {
		return
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	contentHash := contentHash(string(data))

	if !c.content.tryAdd(contentHash) {
		return
	}

	node, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return
	}

	links := extract(node)

	for _, link := range links {
		ch, _ := c.queue.get(link)

		ch <- link
	}

}

func (c *Crawler) run() {
	for _, ch := range c.queue.m {
		go c.work(ch)
	}
}

func (c *Crawler) work(v chan string) {
	for u := range v {
		c.download(u)
	}
}

func contentHash(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])
}
