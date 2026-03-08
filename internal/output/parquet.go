package output

import (
	"fmt"
	"io"

	"github.com/akdavidsson/trawl/internal/extract"
)

// WriteParquet writes extraction results in Parquet format.
func WriteParquet(_ io.Writer, _ *extract.Result) error {
	return fmt.Errorf("parquet output not yet implemented")
}
