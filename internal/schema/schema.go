package schema

import (
	"fmt"
	"strings"
)

// DataType represents the type of a field's data.
type DataType string

const (
	TypeString   DataType = "string"
	TypeInt      DataType = "int"
	TypeFloat    DataType = "float"
	TypeBool     DataType = "bool"
	TypeDate     DataType = "date"
	TypeDatetime DataType = "datetime"
)

// IsValidDataType returns true if dt is a recognized DataType.
func IsValidDataType(dt DataType) bool {
	switch dt {
	case TypeString, TypeInt, TypeFloat, TypeBool, TypeDate, TypeDatetime:
		return true
	}
	return false
}

// Field describes a single field the user wants to extract.
type Field struct {
	Name        string   `json:"name" yaml:"name"`
	Type        DataType `json:"type" yaml:"type"`
	Nullable    bool     `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Schema is the user-defined extraction schema.
type Schema struct {
	Name   string  `json:"name,omitempty" yaml:"name,omitempty"`
	URL    string  `json:"url,omitempty" yaml:"url,omitempty"`
	Fields []Field `json:"fields" yaml:"fields"`
}

// FieldNames returns the list of field names.
func (s *Schema) FieldNames() []string {
	names := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		names[i] = f.Name
	}
	return names
}

// NormalizeName converts a name to a comparable form (lowercase, underscores).
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// Validate checks that the schema is internally consistent.
func (s *Schema) Validate() error {
	if len(s.Fields) == 0 {
		return fmt.Errorf("schema must have at least one field")
	}
	for i, f := range s.Fields {
		if f.Name == "" {
			return fmt.Errorf("field[%d] has empty name", i)
		}
		if f.Type != "" && !IsValidDataType(f.Type) {
			return fmt.Errorf("field %q has invalid type %q", f.Name, f.Type)
		}
	}
	return nil
}
