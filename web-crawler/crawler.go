package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Crawler struct {
	domain  string
	content *Seen
	url     *Seen
	job     chan string
	worker  int
	maximum int
	tasks   sync.WaitGroup
	workers sync.WaitGroup
}

func NewCrawler(domain string, maximum, worker, queueSize int) *Crawler {
	return &Crawler{
		domain:  domain,
		content: NewSeen(),
		url:     NewSeen(),
		maximum: maximum,
		job:     make(chan string, queueSize),
		worker:  worker,
	}
}

func (c *Crawler) download(u string) {
	defer c.tasks.Done()

	u, err := normalizeURL(u)
	if err != nil {
		return
	}

	if !c.url.tryAdd(u) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {

		return
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return
	}

	defer resp.Body.Close()

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

	links := c.extract(node)

	for _, link := range links {
		if c.maximumReached() {
			continue
		}

		c.tasks.Add(1)

		go func() {
			c.job <- link
			fmt.Println(link)
		}()

	}
}

func (c *Crawler) wk() {
	for u := range c.job {
		c.download(u)
	}
}

func (c *Crawler) run(seedUrl string) {
	c.tasks.Add(1)

	c.workers.Go(func() {
		c.download(seedUrl)
	})

	for range c.worker {
		c.workers.Go(func() {
			c.wk()
		})
	}

	c.tasks.Wait()
	close(c.job)

	c.workers.Wait()
}

func contentHash(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])
}
