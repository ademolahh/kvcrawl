# web-crawler

A concurrent web crawler: given a seed URL, it fetches pages, extracts `<a href>` links, and recursively crawls them with a fixed-size worker pool.

Crawling is not restricted to the seed's host — it follows links to other domains too. There is no `robots.txt` handling; only point it at sites you're authorized to crawl.

## Layout

```
web-crawler/
├── main.go          # CLI entrypoint (-seed, -max flags)
├── crawler.go        # worker pool, job queue, download orchestration
├── extract.go         # <a href> extraction from parsed HTML
├── url.go             # link resolution + scheme filtering
└── seen.go            # concurrency-safe dedup sets (visited URLs, content hashes)
```

## How it works

* A fixed pool of worker goroutines pulls URLs off a job channel and downloads them.
* A dispatcher goroutine holds the frontier and feeds the job channel, so workers never block handing off a discovered link and concurrency stays bounded by the pool rather than by the number of links found.
* Each response's content is hashed (SHA-256) and deduplicated, so pages with identical content (e.g. reachable via multiple URLs) are only processed once.
* Discovered links are resolved against the URL of the page they were found on (following redirects), filtered to `http`/`https`, deduplicated, and queued until `-max` unique URLs have been accepted.
* Response bodies are drained and closed on every path, so connections are returned to the pool instead of being leaked on 404s and non-HTML responses.
* A `sync.WaitGroup` tracks outstanding work so the crawler exits once the frontier drains and every in-flight download finishes.

## Running

```sh
go run ./web-crawler -seed https://example.com -max 100
```

| Flag | Description |
| --- | --- |
| `-seed` | Seed URL to start crawling from (required; must be `http` or `https`) |
| `-max` | Maximum number of unique URLs to crawl (`0`, the default, means unlimited) |

Discovered links are printed to stdout as they're queued.

## Tests

```sh
go test -race ./web-crawler
```
