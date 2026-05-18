// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/panz3r/depsclean/internal/model"
)

// resultJSON is the JSON representation of a model.Result with extra human-readable fields.
type resultJSON struct {
	ID              string    `json:"id"`
	Path            string    `json:"path"`
	ProjectPath     string    `json:"project_path"`
	Basename        string    `json:"basename"`
	SizeBytes       int64     `json:"size_bytes"`
	SizeHuman       string    `json:"size_human"`
	LastModified    time.Time `json:"last_modified"`
	LastModifiedISO string    `json:"last_modified_iso"`
	Status          int       `json:"status"`
	PackageManager  string    `json:"package_manager"`
	PackageName     string    `json:"package_name"`
	PackageVersion  string    `json:"package_version"`
	ErrorMsg        string    `json:"error_msg,omitempty"`
}

func toResultJSON(r model.Result) resultJSON {
	return resultJSON{
		ID:              r.ID,
		Path:            r.Path,
		ProjectPath:     r.ProjectPath,
		Basename:        r.Basename,
		SizeBytes:       r.SizeBytes,
		SizeHuman:       formatSize(r.SizeBytes),
		LastModified:    r.LastModified,
		LastModifiedISO: r.LastModified.Format(time.RFC3339),
		Status:          int(r.Status),
		PackageManager:  string(r.PackageManager),
		PackageName:     r.PackageName,
		PackageVersion:  r.PackageVersion,
		ErrorMsg:        r.ErrorMsg,
	}
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes < KB:
		return fmt.Sprintf("%d B", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	case bytes < GB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	}
}

// NDJSONWriter writes one JSON object per line (streaming).
type NDJSONWriter struct {
	w io.Writer
}

// NewNDJSONWriter returns an NDJSONWriter that writes to w.
func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	return &NDJSONWriter{w: w}
}

// Write marshals result as a single JSON line to the underlying writer.
func (j *NDJSONWriter) Write(result model.Result) error {
	data, err := json.Marshal(toResultJSON(result))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(j.w, "%s\n", data)
	return err
}

// JSONWriter collects all results and flushes a full JSON array on Flush.
type JSONWriter struct {
	w       io.Writer
	results []model.Result
}

// NewJSONWriter returns a JSONWriter that writes to w.
func NewJSONWriter(w io.Writer) *JSONWriter {
	return &JSONWriter{w: w}
}

// Add accumulates a result for later flushing.
func (j *JSONWriter) Add(result model.Result) {
	j.results = append(j.results, result)
}

// Flush writes the accumulated results as a JSON array to the underlying writer.
func (j *JSONWriter) Flush() error {
	out := make([]resultJSON, len(j.results))
	for i, r := range j.results {
		out[i] = toResultJSON(r)
	}
	data, err := json.Marshal(out)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(j.w, "%s\n", data)
	return err
}
