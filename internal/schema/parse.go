package schema

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFields parses a comma-separated field string like "name, price, rating"
// into a Schema with string-typed fields (types will be refined by the LLM).
func ParseFields(fields string) (*Schema, error) {
	parts := strings.Split(fields, ",")
	var schema Schema
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		schema.Fields = append(schema.Fields, Field{
			Name: NormalizeName(name),
			Type: TypeString, // default; LLM strategy will refine types
		})
	}
	if len(schema.Fields) == 0 {
		return nil, fmt.Errorf("no fields specified")
	}
	return &schema, nil
}

// ParseYAML reads a YAML schema file and returns the parsed Schema.
func ParseYAML(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing schema YAML: %w", err)
	}

	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	return &s, nil
}
