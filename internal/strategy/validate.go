package strategy

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ValidateAgainstPage checks if a strategy's selectors actually work on the given HTML.
// Returns the number of items found and any issues.
func ValidateAgainstPage(s *ExtractionStrategy, html []byte) (int, []string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return 0, nil, fmt.Errorf("parsing HTML: %w", err)
	}

	items := doc.Find(s.ItemSelector)
	count := items.Length()

	var issues []string
	if count == 0 {
		issues = append(issues, fmt.Sprintf("item_selector %q matched 0 elements", s.ItemSelector))
		return 0, issues, nil
	}

	// Check each field selector against the first item
	first := items.First()
	for _, f := range s.Fields {
		sel := first.Find(f.Selector)
		if sel.Length() == 0 {
			// Try fallbacks
			found := false
			for _, fb := range f.Fallbacks {
				if first.Find(fb).Length() > 0 {
					found = true
					break
				}
			}
			if !found {
				issues = append(issues, fmt.Sprintf("field %q selector %q matched 0 elements", f.Name, f.Selector))
			}
		}
	}

	return count, issues, nil
}
