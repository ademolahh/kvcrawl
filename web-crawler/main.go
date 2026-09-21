package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

func main() {
	seedUrl := flag.String("seed", "", "seed url to start crawling from (required)")
	maximumLink := flag.Int("max", 0, "maximum number of unique urls to crawl (0 means unlimited)")
	flag.Parse()

	if err := validateSeed(*seedUrl); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	c := NewCrawler(*maximumLink, 100, 30)
	c.run(*seedUrl)
}

func validateSeed(seed string) error {
	if seed == "" {
		return fmt.Errorf("-seed is required")
	}

	u, err := url.Parse(seed)
	if err != nil {
		return fmt.Errorf("invalid -seed %q: %w", seed, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("-seed must be an http or https url, got %q", seed)
	}

	if u.Host == "" {
		return fmt.Errorf("-seed has no host: %q", seed)
	}

	return nil
}
