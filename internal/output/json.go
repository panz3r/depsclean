package output

import (
	"io"

	"github.com/panz3r/npclean/internal/model"
)

// JSONWriter writes results as JSON lines to w.
// Implementation is deferred to a later phase.
type JSONWriter struct {
	w io.Writer
}

// NewJSONWriter returns a JSONWriter that writes to w.
func NewJSONWriter(w io.Writer) *JSONWriter {
	return &JSONWriter{w: w}
}

// Write serialises result to the underlying writer.
func (j *JSONWriter) Write(result model.Result) error {
	return nil
}
