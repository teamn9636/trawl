package analyze

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// SimplifyHTML takes raw HTML and returns a cleaned, minified version
// suitable for sending to an LLM. It removes scripts, styles, comments,
// and non-structural elements while preserving the DOM structure.
func SimplifyHTML(html []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return "", err
	}

	// Remove elements that add noise for the LLM
	doc.Find("script, style, noscript, iframe, svg, link, meta, head > *:not(title)").Remove()

	// Remove comments and hidden elements
	doc.Find("[style*='display:none'], [style*='display: none'], [hidden]").Remove()

	// Remove empty elements that add no value
	doc.Find("br, hr").Remove()

	// Strip data attributes and event handlers to reduce token count
	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		var toRemove []string
		for _, attr := range s.Get(0).Attr {
			if strings.HasPrefix(attr.Key, "data-") ||
				strings.HasPrefix(attr.Key, "on") ||
				attr.Key == "style" ||
				attr.Key == "tabindex" ||
				attr.Key == "aria-hidden" {
				toRemove = append(toRemove, attr.Key)
			}
		}
		for _, key := range toRemove {
			s.RemoveAttr(key)
		}
	})

	body := doc.Find("body")
	if body.Length() == 0 {
		body = doc.Selection
	}

	result, err := body.Html()
	if err != nil {
		return "", err
	}

	return collapseWhitespace(result), nil
}

// collapseWhitespace reduces runs of whitespace to single spaces and trims lines.
func collapseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, "\n")
}
