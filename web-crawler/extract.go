package main

import (
	"golang.org/x/net/html"
)

func extract(n *html.Node) []string {
	var out []string

	var rec (func(node *html.Node))

	rec = func(node *html.Node) {}

	rec(n)

	return out
}
