package schema

import (
	"testing"
)

func TestParseFields(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		wantErr  bool
	}{
		{"name, price, rating", []string{"name", "price", "rating"}, false},
		{"name,price,rating", []string{"name", "price", "rating"}, false},
		{" name , price ", []string{"name", "price"}, false},
		{"Product Name, Sale Price", []string{"product_name", "sale_price"}, false},
		{"", nil, true},
		{",,,", nil, true},
	}

	for _, tt := range tests {
		s, err := ParseFields(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseFields(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseFields(%q) error: %v", tt.input, err)
			continue
		}
		if len(s.Fields) != len(tt.expected) {
			t.Errorf("ParseFields(%q) got %d fields, want %d", tt.input, len(s.Fields), len(tt.expected))
			continue
		}
		for i, f := range s.Fields {
			if f.Name != tt.expected[i] {
				t.Errorf("ParseFields(%q) field[%d] = %q, want %q", tt.input, i, f.Name, tt.expected[i])
			}
		}
	}
}

func TestParseYAML(t *testing.T) {
	s, err := ParseYAML("../../examples/products.yaml")
	if err != nil {
		t.Fatalf("ParseYAML error: %v", err)
	}
	if len(s.Fields) != 5 {
		t.Errorf("got %d fields, want 5", len(s.Fields))
	}
	if s.Name != "product_listing" {
		t.Errorf("got name %q, want %q", s.Name, "product_listing")
	}
	if s.Fields[1].Type != TypeFloat {
		t.Errorf("price field type = %q, want %q", s.Fields[1].Type, TypeFloat)
	}
	if !s.Fields[4].Nullable {
		t.Error("rating field should be nullable")
	}
}
