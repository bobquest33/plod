package html

import (
	"io"
	"net/url"

	"github.com/sjwhitworth/plod/domain"
	"golang.org/x/net/html"
)

// Code from github.com/jackdanger/collectlinks
// Copyright Jack Danger Canty
// Pulled into own repo to change function signatures

func ParseLinks(httpBody io.Reader) []domain.URL {
	links := make([]domain.URL, 0)
	page := html.NewTokenizer(httpBody)
	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					links = append(links, domain.URL(attr.Val))
				}
			}
		}
	}
}

func FixURL(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}
