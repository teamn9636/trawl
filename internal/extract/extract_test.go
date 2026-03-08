package extract

import (
	"os"
	"testing"

	"github.com/akdavidsson/trawl/internal/strategy"
)

func TestApply(t *testing.T) {
	html, err := os.ReadFile("../../testdata/sample.html")
	if err != nil {
		t.Fatalf("reading sample HTML: %v", err)
	}

	strat := &strategy.ExtractionStrategy{
		ItemSelector: "div.product-card",
		Fields: []strategy.FieldMapping{
			{Name: "name", Selector: "h2.product-title", Attribute: "text", Type: "string"},
			{Name: "price", Selector: "span.price", Attribute: "text", Transform: "parse_price", Type: "float"},
			{Name: "rating", Selector: "span.rating", Attribute: "text", Type: "float"},
			{Name: "in_stock", Selector: "span.stock", Attribute: "text", Type: "bool"},
			{Name: "url", Selector: "a.product-link", Attribute: "href", Type: "string"},
		},
	}

	result, err := Apply(strat, html)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Records) != 4 {
		t.Fatalf("got %d records, want 4", len(result.Records))
	}

	// Check first record
	rec := result.Records[0]
	if rec["name"] != "Widget Alpha" {
		t.Errorf("name = %v, want %q", rec["name"], "Widget Alpha")
	}
	if rec["price"] != float64(29.99) {
		t.Errorf("price = %v, want 29.99", rec["price"])
	}
	if rec["rating"] != float64(4.5) {
		t.Errorf("rating = %v, want 4.5", rec["rating"])
	}
	if rec["in_stock"] != true {
		t.Errorf("in_stock = %v, want true", rec["in_stock"])
	}
	if rec["url"] != "/products/widget-alpha" {
		t.Errorf("url = %v, want %q", rec["url"], "/products/widget-alpha")
	}

	// Check out-of-stock item
	rec3 := result.Records[2]
	if rec3["in_stock"] != false {
		t.Errorf("record 3 in_stock = %v, want false", rec3["in_stock"])
	}
}

func TestApplyWithFallbacks(t *testing.T) {
	html := []byte(`<html><body>
		<div class="item"><span class="alt-name">Test</span><span class="price">$10</span></div>
	</body></html>`)

	strat := &strategy.ExtractionStrategy{
		ItemSelector: "div.item",
		Fields: []strategy.FieldMapping{
			{
				Name:      "name",
				Selector:  "h2.title",
				Attribute: "text",
				Type:      "string",
				Fallbacks: []string{"span.alt-name", "span.name"},
			},
		},
	}

	result, err := Apply(strat, html)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(result.Records))
	}
	if result.Records[0]["name"] != "Test" {
		t.Errorf("name = %v, want %q", result.Records[0]["name"], "Test")
	}
}

func TestTransformValue(t *testing.T) {
	tests := []struct {
		raw       string
		transform string
		typ       string
		want      any
	}{
		{"$29.99", "parse_price", "float", float64(29.99)},
		{"1,234.56", "", "float", float64(1234.56)},
		{"42", "", "int", int64(42)},
		{"true", "", "bool", true},
		{"In Stock", "", "bool", true},
		{"Out of Stock", "", "bool", false},
		{"hello", "", "string", "hello"},
		{"", "", "string", nil},
	}

	for _, tt := range tests {
		got, warn := TransformValue(tt.raw, tt.transform, tt.typ)
		if got != tt.want {
			t.Errorf("TransformValue(%q, %q, %q) = %v (%T), want %v (%T) [warn: %s]",
				tt.raw, tt.transform, tt.typ, got, got, tt.want, tt.want, warn)
		}
	}
}

func TestComputeHealth(t *testing.T) {
	result := &Result{
		Fields: []string{"name", "price", "rating"},
		Records: []Record{
			{"name": "A", "price": float64(10), "rating": float64(4.5)},
			{"name": "B", "price": float64(20), "rating": nil},
			{"name": "C", "price": nil, "rating": nil},
		},
	}

	health := ComputeHealth(result)
	if health.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d, want 3", health.TotalRecords)
	}
	if health.TotalFields != 9 {
		t.Errorf("TotalFields = %d, want 9", health.TotalFields)
	}
	if health.PopulatedFields != 6 {
		t.Errorf("PopulatedFields = %d, want 6", health.PopulatedFields)
	}
	// 6/9 = 66.7% => should need re-inference at 70% threshold
	if !health.NeedsReInference(70) {
		t.Error("should need re-inference at 70% threshold")
	}
	if health.NeedsReInference(50) {
		t.Error("should not need re-inference at 50% threshold")
	}
}
