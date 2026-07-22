package main

import "flag"

const CRAWL string = "https://ademola.xyz"

func main() {
	var seedUrl = flag.String("seed", "", "seed url")
	var maximumLink = flag.Int("max", 0, "maximum link to crawl")
	flag.Parse()

	c := NewCrawler(*seedUrl, *maximumLink, 100, 30)
	c.run(*seedUrl)
}
