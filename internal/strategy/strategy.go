package strategy

import "encoding/json"

// ExtractionStrategy is the LLM-derived plan for extracting data from a page.
type ExtractionStrategy struct {
	SitePattern       string         `json:"site_pattern"`
	ContainerSelector string         `json:"container_selector,omitempty"` // scopes item_selector to a specific page section
	ItemSelector      string         `json:"item_selector"`
	Fields            []FieldMapping `json:"fields"`
	Pagination        *PaginationRule `json:"pagination,omitempty"`
	Confidence        float64        `json:"confidence"`
	Fingerprint       string         `json:"fingerprint"`
}

// FieldMapping describes how to extract a single field from an item element.
type FieldMapping struct {
	Name      string   `json:"name"`
	Selector  string   `json:"selector"`
	Attribute string   `json:"attribute"` // "text", "href", "src", or any HTML attribute
	Transform string   `json:"transform,omitempty"` // "trim", "parse_price", "parse_date"
	Type      string   `json:"type"`
	Fallbacks []string `json:"fallbacks,omitempty"`
}

// PaginationRule describes how to navigate between pages.
type PaginationRule struct {
	Type       string `json:"type"`        // "next_link", "url_increment", "load_more", "infinite_scroll"
	Selector   string `json:"selector"`
	URLPattern string `json:"url_pattern,omitempty"`
	HasMore    string `json:"has_more,omitempty"`
}

// toJSON returns the strategy as a JSON string (for multi-turn LLM conversations).
func (s *ExtractionStrategy) toJSON() string {
	b, err := json.Marshal(s)
	if err != nil {
		return "{}"
	}
	return string(b)
}
