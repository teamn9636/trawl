package analyze

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Fingerprint computes a structural hash of an HTML page.
// It captures the DOM tree structure (tag hierarchy) without content,
// so that two pages with the same layout but different data produce
// the same fingerprint.
func Fingerprint(html []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return "", err
	}

	body := doc.Find("body")
	if body.Length() == 0 {
		body = doc.Selection
	}

	var sb strings.Builder
	buildStructure(body, &sb, 0)

	hash := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", hash[:8]), nil
}

// buildStructure recursively builds a string representing the DOM structure.
// Only tag names and key structural attributes (class, id, role) are included.
func buildStructure(s *goquery.Selection, sb *strings.Builder, depth int) {
	if depth > 20 {
		return // avoid excessive depth
	}
	s.Children().Each(func(_ int, child *goquery.Selection) {
		node := child.Get(0)
		tag := node.Data

		// Skip non-structural elements
		if tag == "script" || tag == "style" || tag == "noscript" {
			return
		}

		sb.WriteString(tag)
		if cls, ok := child.Attr("class"); ok {
			sb.WriteString(".")
			sb.WriteString(cls)
		}
		if id, ok := child.Attr("id"); ok {
			sb.WriteString("#")
			sb.WriteString(id)
		}
		sb.WriteString("{")
		buildStructure(child, sb, depth+1)
		sb.WriteString("}")
	})
}
