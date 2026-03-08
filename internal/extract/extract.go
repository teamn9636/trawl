package extract

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/akdavidsson/trawl/internal/strategy"
)

// Record is a single extracted data row.
type Record = map[string]any

// Result holds the extraction output.
type Result struct {
	Records  []Record
	Fields   []string // ordered field names
	Warnings []string
}

// Apply applies an extraction strategy to an HTML page and returns structured records.
func Apply(strat *strategy.ExtractionStrategy, html []byte) (*Result, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	items := doc.Find(strat.ItemSelector)
	if items.Length() == 0 {
		return nil, fmt.Errorf("item selector %q matched 0 elements", strat.ItemSelector)
	}

	fieldNames := make([]string, len(strat.Fields))
	for i, f := range strat.Fields {
		fieldNames[i] = f.Name
	}

	result := &Result{Fields: fieldNames}

	items.Each(func(_ int, item *goquery.Selection) {
		rec := make(Record, len(strat.Fields))
		for _, field := range strat.Fields {
			raw := extractField(item, field)
			val, warn := TransformValue(raw, field.Transform, field.Type)
			if warn != "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("field %q: %s", field.Name, warn))
			}
			rec[field.Name] = val
		}
		result.Records = append(result.Records, rec)
	})

	return result, nil
}

// extractField extracts a single field value from an item using selectors.
func extractField(item *goquery.Selection, field strategy.FieldMapping) string {
	// Try primary selector
	val := extractFromSelector(item, field.Selector, field.Attribute)
	if val != "" {
		return val
	}

	// Try fallback selectors
	for _, fb := range field.Fallbacks {
		val = extractFromSelector(item, fb, field.Attribute)
		if val != "" {
			return val
		}
	}

	return ""
}

// extractFromSelector applies a CSS selector and extracts the value.
func extractFromSelector(item *goquery.Selection, selector, attribute string) string {
	var sel *goquery.Selection
	if selector == "" || selector == "." {
		sel = item
	} else {
		sel = item.Find(selector)
	}

	if sel.Length() == 0 {
		return ""
	}

	first := sel.First()
	switch strings.ToLower(attribute) {
	case "text", "":
		return strings.TrimSpace(first.Text())
	case "html":
		h, _ := first.Html()
		return strings.TrimSpace(h)
	default:
		val, exists := first.Attr(attribute)
		if !exists {
			return ""
		}
		return strings.TrimSpace(val)
	}
}
