package output

import (
	"encoding/json"
	"io"

	"github.com/akdavidsson/trawl/internal/extract"
)

// WriteJSONL writes the extraction result as newline-delimited JSON.
func WriteJSONL(w io.Writer, result *extract.Result) error {
	enc := json.NewEncoder(w)
	for _, rec := range result.Records {
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}
