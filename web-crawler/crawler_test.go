package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"slices"
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
			c := NewCrawler(100, 100, 10)
			node := parseHTML(t, tt.html)

			got := c.extract("https://example.com", node)

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

	c := NewCrawler(100, 10, 10)

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

// Relative hrefs must resolve against the page they appear on. Resolving them
// against the seed sends every link on a nested page to the wrong path.
func TestRelativeLinksResolveAgainstPage(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	record := func(p string) {
		mu.Lock()
		seen = append(seen, p)
		mu.Unlock()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		record("/")
		fmt.Fprint(w, `<a href="/deep/page">deep</a>`)
	})
	mux.HandleFunc("/deep/page", func(w http.ResponseWriter, r *http.Request) {
		record("/deep/page")
		fmt.Fprint(w, `<a href="sibling">sibling</a><a href="../top">top</a>`)
	})
	mux.HandleFunc("/deep/sibling", func(w http.ResponseWriter, r *http.Request) {
		record("/deep/sibling")
		fmt.Fprint(w, `sibling body`)
	})
	mux.HandleFunc("/top", func(w http.ResponseWriter, r *http.Request) {
		record("/top")
		fmt.Fprint(w, `top body`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewCrawler(100, 5, 10)
	c.run(srv.URL)

	mu.Lock()
	defer mu.Unlock()
	for _, want := range []string{"/", "/deep/page", "/deep/sibling", "/top"} {
		if !slices.Contains(seen, want) {
			t.Errorf("never requested %q; requested %v", want, seen)
		}
	}
	if slices.Contains(seen, "/sibling") {
		t.Errorf("requested /sibling: relative link resolved against the seed, not the page")
	}
}

// Links are resolved against the post-redirect URL, not the one we asked for.
func TestLinksResolveAgainstRedirectTarget(t *testing.T) {
	var mu sync.Mutex
	var seen []string

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<a href="/start">start</a>`)
	})
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/moved/here", http.StatusFound)
	})
	mux.HandleFunc("/moved/here", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<a href="next">next</a>`)
	})
	mux.HandleFunc("/moved/next", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, "/moved/next")
		mu.Unlock()
		fmt.Fprint(w, `arrived`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewCrawler(100, 5, 10)
	c.run(srv.URL)

	mu.Lock()
	defer mu.Unlock()
	if !slices.Contains(seen, "/moved/next") {
		t.Errorf("never requested /moved/next: link resolved against the pre-redirect URL")
	}
}

// Every early return in download must still close the response body, or the
// connection is never returned to the pool.
func TestResponseBodyAlwaysClosed(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"non-200", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusNotFound)
		}},
		{"non-html content type", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"a":1}`)
		}},
		{"html", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<p>hi</p>`)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			conns := 0

			srv := httptest.NewUnstartedServer(tt.handler)
			srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
				if s == http.StateNew {
					mu.Lock()
					conns++
					mu.Unlock()
				}
			}
			srv.Start()
			defer srv.Close()

			const requests = 10
			c := NewCrawler(100, 1, 10)
			for i := range requests {
				c.tasks.Add(1)
				c.download(fmt.Sprintf("%s/p%d", srv.URL, i))
			}

			mu.Lock()
			defer mu.Unlock()
			// A closed body lets the transport reuse one connection; a leaked
			// one forces a fresh connection per request.
			if conns > 2 {
				t.Errorf("%d sequential requests opened %d connections, want <= 2 (response bodies are leaking)", requests, conns)
			}
		})
	}
}

// -max bounds the URLs accepted, and is applied when a link is queued rather
// than when it is downloaded.
func TestMaximumIsEnforced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Unique body per path so content dedup never prunes the fan-out.
		fmt.Fprintf(w, "<p>%s</p>", r.URL.Path)
		for i := range 50 {
			fmt.Fprintf(w, `<a href="%s/x%d">x</a>`, r.URL.Path, i)
		}
	}))
	defer srv.Close()

	const max = 10
	c := NewCrawler(max, 20, 10)
	c.run(srv.URL)

	if got := c.url.Size(); got != max {
		t.Errorf("accepted %d urls, want exactly %d", got, max)
	}
}

// A -max of 0 (the flag default) means unlimited, not "crawl nothing".
func TestMaximumZeroMeansUnlimited(t *testing.T) {
	var mu sync.Mutex
	fetched := 0

	mux := http.NewServeMux()
	for _, p := range []string{"/", "/a", "/b"} {
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			fetched++
			mu.Unlock()
			fmt.Fprintf(w, `<p>%s</p><a href="/a">A</a><a href="/b">B</a>`, r.URL.Path)
		})
	}

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewCrawler(0, 5, 10)
	c.run(srv.URL)

	mu.Lock()
	defer mu.Unlock()
	if fetched != 3 {
		t.Errorf("fetched %d pages with -max 0, want 3 (0 should mean unlimited)", fetched)
	}
}

// Discovered links must not each get their own goroutine; the pool size is
// what bounds concurrency.
func TestGoroutinesBoundedByWorkerPool(t *testing.T) {
	const links = 2000

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fmt.Fprintf(w, "<p>%s</p>", r.URL.Path)
			return
		}
		for i := range links {
			fmt.Fprintf(w, `<a href="/p%d">x</a>`, i)
		}
	}))
	defer srv.Close()

	before := runtime.NumGoroutine()

	var peak int64
	stop := make(chan struct{})
	var watcher sync.WaitGroup
	watcher.Go(func() {
		for {
			select {
			case <-stop:
				return
			default:
				if n := int64(runtime.NumGoroutine()); n > peak {
					peak = n
				}
				time.Sleep(time.Millisecond)
			}
		}
	})

	const workers = 4
	c := NewCrawler(links+1, workers, 5)
	c.run(srv.URL)
	close(stop)
	watcher.Wait()

	// Generous headroom for the pool, the dispatcher, the test server's own
	// per-connection goroutines and the watcher — but far below one goroutine
	// per discovered link.
	if limit := int64(before + workers + 100); peak > limit {
		t.Errorf("peak goroutines %d for %d links, want <= %d (a goroutine per link is unbounded)", peak, links, limit)
	}
}

func TestValidateSeed(t *testing.T) {
	tests := []struct {
		name    string
		seed    string
		wantErr bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http with path", "http://example.com/a/b", false},
		{"empty", "", true},
		{"no scheme", "example.com", true},
		{"bare word", "not a url", true},
		{"unsupported scheme", "ftp://example.com", true},
		{"no host", "http://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSeed(tt.seed)
			if tt.wantErr && err == nil {
				t.Errorf("validateSeed(%q) = nil, want an error", tt.seed)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateSeed(%q) = %v, want nil", tt.seed, err)
			}
		})
	}
}
