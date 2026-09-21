package main

import (
	"net/url"
)

func resolveURL(base, path string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return b.ResolveReference(ref).String(), nil
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

func process(base, path string) (string, bool) {
	resolved, err := resolveURL(base, path)
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
