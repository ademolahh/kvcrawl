package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func parseHTML(t *testing.T, s string) *html.Node {
	t.Helper()
	node, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}
	return node
}

func TestExtraction(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []string
	}{
		{
			name: "relative link resolved against domain",
			html: `<a href="/about">About</a>`,
			want: []string{"https://example.com/about"},
		},
		{
			name: "absolute same-domain link kept as-is",
			html: `<a href="https://example.com/contact">Contact</a>`,
			want: []string{"https://example.com/contact"},
		},
		{
			name: "off-domain link is followed, not filtered",
			html: `<a href="https://other-site.org/page">Other</a>`,
			want: []string{"https://other-site.org/page"},
		},
		{
			name: "non-http(s) schemes are dropped",
			html: `<a href="mailto:hi@example.com">Mail</a><a href="javascript:void(0)">JS</a><a href="/ok">OK</a>`,
			want: []string{"https://example.com/ok"},
		},
		{
			name: "anchor with no href is skipped",
			html: `<a name="top">Top</a><a href="/ok">OK</a>`,
			want: []string{"https://example.com/ok"},
		},
		{
			name: "links nested inside other elements are found",
			html: `<div><ul><li><a href="/one">One</a></li></ul></div><footer><a href="/two">Two</a></footer>`,
			want: []string{"https://example.com/one", "https://example.com/two"},
		},
		{
			name: "no anchors returns nil",
			html: `<div>no links here</div>`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCrawler("https://example.com", 100, 100, 10)
			node := parseHTML(t, tt.html)

			got := c.extract(node)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	var mu sync.Mutex
	hits := map[string]int{}
	record := func(path string) {
		mu.Lock()
		hits[path]++
		mu.Unlock()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		record("/")
		fmt.Fprint(w, `<a href="/a">A</a><a href="/b">B</a>`)
	})
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		record("/a")
		fmt.Fprint(w, `<a href="/c">C</a>`)
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		record("/b")
		fmt.Fprint(w, `<a href="/">Home</a>`) // links back to the seed
	})
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		record("/c")
		fmt.Fprint(w, `no links here`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewCrawler(srv.URL, 100, 10, 10)

	done := make(chan struct{})
	go func() {
		c.run(srv.URL)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("run() did not terminate within 5s — possible deadlock")
	}

	mu.Lock()
	defer mu.Unlock()
	for _, path := range []string{"/", "/a", "/b", "/c"} {
		if hits[path] != 1 {
			t.Errorf("hits[%q] = %d, want 1 (dedup should prevent re-crawling a seen URL)", path, hits[path])
		}
	}
}
