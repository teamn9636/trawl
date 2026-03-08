package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/akdavidsson/trawl/internal/extract"
)

func testResult() *extract.Result {
	return &extract.Result{
		Fields: []string{"name", "price", "in_stock"},
		Records: []extract.Record{
			{"name": "Widget A", "price": float64(29.99), "in_stock": true},
			{"name": "Widget B", "price": float64(49.99), "in_stock": false},
		},
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, testResult()); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}

	var records []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("got %d records, want 2", len(records))
	}
	if records[0]["name"] != "Widget A" {
		t.Errorf("first name = %v, want %q", records[0]["name"], "Widget A")
	}
}

func TestWriteJSONL(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSONL(&buf, testResult()); err != nil {
		t.Fatalf("WriteJSONL error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("got %d lines, want 2", len(lines))
	}

	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &rec); err != nil {
		t.Fatalf("invalid JSONL line: %v", err)
	}
	if rec["name"] != "Widget A" {
		t.Errorf("first name = %v, want %q", rec["name"], "Widget A")
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, testResult()); err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Errorf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "name,price,in_stock" {
		t.Errorf("header = %q, want %q", lines[0], "name,price,in_stock")
	}
	if !strings.Contains(lines[1], "Widget A") {
		t.Error("first data row should contain 'Widget A'")
	}
}
