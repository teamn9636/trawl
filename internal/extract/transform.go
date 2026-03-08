package extract

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var priceRegex = regexp.MustCompile(`[\d,]+\.?\d*`)

// TransformValue applies an optional transform and type coercion to a raw string.
func TransformValue(raw, transform, targetType string) (any, string) {
	if raw == "" {
		return nil, ""
	}

	// Apply transform first
	switch strings.ToLower(transform) {
	case "trim":
		raw = strings.TrimSpace(raw)
	case "parse_price":
		raw = parsePrice(raw)
	case "parse_date":
		return parseDate(raw)
	case "parse_int":
		return parseInt(raw)
	case "parse_float":
		return parseFloat(raw)
	}

	// Then coerce to target type
	switch strings.ToLower(targetType) {
	case "int":
		return parseInt(raw)
	case "float":
		return parseFloat(raw)
	case "bool":
		return parseBool(raw)
	case "date", "datetime":
		return parseDate(raw)
	default:
		return raw, ""
	}
}

func parsePrice(raw string) string {
	match := priceRegex.FindString(raw)
	if match == "" {
		return raw
	}
	return strings.ReplaceAll(match, ",", "")
}

func parseInt(raw string) (any, string) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	v, err := strconv.ParseInt(cleaned, 10, 64)
	if err != nil {
		// Try parsing as float and truncating
		f, ferr := strconv.ParseFloat(cleaned, 64)
		if ferr != nil {
			return raw, fmt.Sprintf("cannot parse %q as int", raw)
		}
		return int64(f), ""
	}
	return v, ""
}

func parseFloat(raw string) (any, string) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	v, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return raw, fmt.Sprintf("cannot parse %q as float", raw)
	}
	return v, ""
}

func parseBool(raw string) (any, string) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "yes", "1", "in stock", "available":
		return true, ""
	case "false", "no", "0", "out of stock", "unavailable":
		return false, ""
	}
	return raw, fmt.Sprintf("cannot parse %q as bool", raw)
}

var dateFmts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02",
	"01/02/2006",
	"Jan 2, 2006",
	"January 2, 2006",
	"02-Jan-2006",
	"02 Jan 2006",
}

func parseDate(raw string) (any, string) {
	trimmed := strings.TrimSpace(raw)
	for _, fmt := range dateFmts {
		if t, err := time.Parse(fmt, trimmed); err == nil {
			return t, ""
		}
	}
	return raw, fmt.Sprintf("cannot parse %q as date", raw)
}
