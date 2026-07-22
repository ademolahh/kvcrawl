package main

import (
	"golang.org/x/net/html"
)

func (c *Crawler) extract(n *html.Node) []string {
	var out []string

	var rec (func(node *html.Node))

	rec = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {

				if attr.Key != "href" {
					continue
				}

				link, ok := c.process(attr.Val)
				if !ok {
					continue
				}

				out = append(out, link)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}

	rec(n)

	return out
}
