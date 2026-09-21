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
	content *Seen
	url     *Seen
	client  *http.Client

	discovered chan string
	job        chan string

	worker  int
	maximum int
	tasks   sync.WaitGroup
	workers sync.WaitGroup
}

func NewCrawler(maximum, worker, queueSize int) *Crawler {
	return &Crawler{
		content:    NewSeen(),
		url:        NewSeen(),
		client:     &http.Client{},
		discovered: make(chan string, queueSize),
		job:        make(chan string, queueSize),
		maximum:    maximum,
		worker:     worker,
	}
}

func (c *Crawler) enqueue(rawURL string) {
	u, err := normalizeURL(rawURL)
	if err != nil {
		return
	}

	if !c.url.tryAddBounded(u, c.maximum) {
		return
	}

	c.tasks.Add(1)
	c.discovered <- u
	fmt.Println(u)
}

func (c *Crawler) download(u string) {
	defer c.tasks.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer drainAndClose(resp)

	if resp.StatusCode != http.StatusOK {
		return
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if !c.content.tryAdd(contentHash(data)) {
		return
	}

	node, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return
	}

	for _, link := range c.extract(resp.Request.URL.String(), node) {
		c.enqueue(link)
	}
}

func (c *Crawler) wk() {
	for u := range c.job {
		c.download(u)
	}
}

func (c *Crawler) dispatch() {
	defer close(c.job)

	var queue []string
	in := c.discovered

	for in != nil || len(queue) > 0 {

		var out chan<- string
		var next string
		if len(queue) > 0 {
			out, next = c.job, queue[0]
		}

		select {
		case link, ok := <-in:
			if !ok {
				in = nil
				continue
			}
			queue = append(queue, link)
		case out <- next:
			queue = queue[1:]
		}
	}
}

func (c *Crawler) run(seedUrl string) {
	var dispatcher sync.WaitGroup
	dispatcher.Go(c.dispatch)

	for range c.worker {
		c.workers.Go(c.wk)
	}

	c.enqueue(seedUrl)

	c.tasks.Wait()
	close(c.discovered)

	dispatcher.Wait()
	c.workers.Wait()
}

const maxDrain = 64 << 10

func drainAndClose(resp *http.Response) {
	io.CopyN(io.Discard, resp.Body, maxDrain)
	resp.Body.Close()
}

func contentHash(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}
