package main

import (
	"net/url"
)

func (c *Crawler) resolveUrl(path string) (string, error) {
	base, err := url.Parse(c.domain)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	resolved := base.ResolveReference(ref)

	return resolved.String(), nil
}

func normalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}

	return u.String(), nil
}

func (c *Crawler) process(u string) (string, bool) {
	resolved, err := c.resolveUrl(u)
	if err != nil {
		return "", false
	}

	ur, err := url.Parse(resolved)
	if err != nil {
		return "", false
	}

	if ur.Scheme != "https" && ur.Scheme != "http" {
		return "", false
	}

	return resolved, true

}
