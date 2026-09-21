package main

import (
	"golang.org/x/net/html"
)

func (c *Crawler) extract(base string, n *html.Node) []string {
	var out []string

	var rec (func(node *html.Node))

	rec = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {

				if attr.Key != "href" {
					continue
				}

				link, ok := process(base, attr.Val)
				if !ok {
					continue
				}

				out = append(out, link)
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			rec(child)
		}
	}

	rec(n)

	return out
}
