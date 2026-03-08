package output

import (
	"encoding/json"
	"io"

	"github.com/akdavidsson/trawl/internal/extract"
)

// WriteJSON writes the extraction result as an indented JSON array.
func WriteJSON(w io.Writer, result *extract.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result.Records)
}
