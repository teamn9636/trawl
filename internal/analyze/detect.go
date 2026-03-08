package analyze

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CandidateRegion represents a part of the page that likely contains
// repeating data (tables, lists, grids of similar elements).
type CandidateRegion struct {
	HTML        string // simplified HTML of the region
	Tag         string // root element tag (table, ul, div, etc.)
	Selector    string // CSS selector to find this region
	ItemCount   int    // number of repeated items found
	HasHeaders  bool   // whether the region has obvious headers (th, thead)
	Context     string // nearby heading or caption text
}

// DetectCandidateRegions scans an HTML document for regions that likely
// contain structured, repeating data. Returns regions sorted by likelihood.
func DetectCandidateRegions(html []byte) ([]CandidateRegion, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return nil, err
	}

	var regions []CandidateRegion

	// 1. Find all <table> elements (highest confidence)
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		rows := s.Find("tr").Length()
		if rows < 2 {
			return
		}
		region := CandidateRegion{
			Tag:        "table",
			Selector:   buildSelector(s, i),
			ItemCount:  rows,
			HasHeaders: s.Find("thead, th").Length() > 0,
			Context:    findContext(s),
		}
		h, _ := goquery.OuterHtml(s)
		region.HTML = h
		regions = append(regions, region)
	})

	// 2. Find repeated list items
	doc.Find("ul, ol").Each(func(i int, s *goquery.Selection) {
		items := s.Find("li").Length()
		if items < 3 {
			return
		}
		region := CandidateRegion{
			Tag:       s.Get(0).Data,
			Selector:  buildSelector(s, i),
			ItemCount: items,
			Context:   findContext(s),
		}
		h, _ := goquery.OuterHtml(s)
		region.HTML = h
		regions = append(regions, region)
	})

	// 3. Find divs/sections with repeated child patterns
	doc.Find("div, section, main").Each(func(i int, s *goquery.Selection) {
		children := s.Children()
		if children.Length() < 3 {
			return
		}
		// Check if children share the same tag+class pattern
		pattern := childPattern(children.First())
		if pattern == "" {
			return
		}
		matching := 0
		children.Each(func(_ int, c *goquery.Selection) {
			if childPattern(c) == pattern {
				matching++
			}
		})
		if matching < 3 || float64(matching)/float64(children.Length()) < 0.5 {
			return
		}
		region := CandidateRegion{
			Tag:       "div",
			Selector:  buildSelector(s, i),
			ItemCount: matching,
			Context:   findContext(s),
		}
		h, _ := goquery.OuterHtml(s)
		region.HTML = h
		regions = append(regions, region)
	})

	return regions, nil
}

// buildSelector creates a CSS selector for a goquery selection.
func buildSelector(s *goquery.Selection, index int) string {
	node := s.Get(0)
	sel := node.Data

	if id, ok := s.Attr("id"); ok && id != "" {
		return sel + "#" + id
	}
	if cls, ok := s.Attr("class"); ok && cls != "" {
		classes := strings.Fields(cls)
		if len(classes) > 0 {
			return sel + "." + strings.Join(classes, ".")
		}
	}
	// Fallback: use nth-of-type
	return fmt.Sprintf("%s:nth-of-type(%d)", sel, index+1)
}

// childPattern returns a string representing the tag+class of a selection.
func childPattern(s *goquery.Selection) string {
	if s.Length() == 0 {
		return ""
	}
	node := s.Get(0)
	pattern := node.Data
	if cls, ok := s.Attr("class"); ok {
		pattern += "." + cls
	}
	return pattern
}

// findContext looks for a nearby heading or caption that describes a region.
func findContext(s *goquery.Selection) string {
	// Check for caption inside
	if cap := s.Find("caption").First(); cap.Length() > 0 {
		return strings.TrimSpace(cap.Text())
	}

	// Check preceding sibling headings
	prev := s.Prev()
	for i := 0; i < 3 && prev.Length() > 0; i++ {
		tag := prev.Get(0).Data
		if tag == "h1" || tag == "h2" || tag == "h3" || tag == "h4" || tag == "h5" || tag == "h6" {
			return strings.TrimSpace(prev.Text())
		}
		prev = prev.Prev()
	}

	// Check parent for aria-label
	if label, ok := s.Parent().Attr("aria-label"); ok {
		return label
	}

	return ""
}
