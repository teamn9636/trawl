package crawl

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/akdavidsson/trawl/internal/strategy"
)

// FindNextPageURL extracts the next page URL from the current page using the pagination rule.
func FindNextPageURL(html []byte, baseURL string, rule *strategy.PaginationRule) (string, error) {
	if rule == nil || rule.Type == "none" {
		return "", nil
	}

	switch rule.Type {
	case "next_link":
		return findNextLink(html, baseURL, rule)
	case "url_increment":
		return "", fmt.Errorf("url_increment pagination not yet implemented")
	default:
		return "", fmt.Errorf("unsupported pagination type: %s", rule.Type)
	}
}

func findNextLink(html []byte, baseURL string, rule *strategy.PaginationRule) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return "", err
	}

	// Check if there are more pages
	if rule.HasMore != "" {
		if doc.Find(rule.HasMore).Length() == 0 {
			return "", nil // no more pages
		}
	}

	sel := doc.Find(rule.Selector)
	if sel.Length() == 0 {
		return "", nil // no next link found
	}

	href, exists := sel.First().Attr("href")
	if !exists || href == "" {
		return "", nil
	}

	// Resolve relative URLs
	if strings.HasPrefix(href, "/") {
		// Extract scheme+host from baseURL
		parts := strings.SplitN(baseURL, "//", 2)
		if len(parts) == 2 {
			hostEnd := strings.Index(parts[1], "/")
			if hostEnd == -1 {
				href = parts[0] + "//" + parts[1] + href
			} else {
				href = parts[0] + "//" + parts[1][:hostEnd] + href
			}
		}
	} else if !strings.HasPrefix(href, "http") {
		// Relative path
		lastSlash := strings.LastIndex(baseURL, "/")
		if lastSlash > 0 {
			href = baseURL[:lastSlash+1] + href
		}
	}

	return href, nil
}
