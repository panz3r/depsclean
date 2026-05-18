// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/panz3r/depsclean/internal/model"
)

func makeResult(id, path string, size int64) model.Result {
	return model.Result{
		ID:           id,
		Path:         path,
		Basename:     path,
		SizeBytes:    size,
		LastModified: time.Now(),
		Status:       model.StatusReady,
	}
}

func TestNDJSONWriter_SingleResult(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)
	r := makeResult("1", "/path/to/node_modules", 1024)
	if err := w.Write(r); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	line := strings.TrimSpace(buf.String())
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatalf("invalid JSON: %v\nline: %s", err, line)
	}
	if out["id"] != "1" {
		t.Errorf("expected id=1, got %v", out["id"])
	}
	if out["size_human"] != "1.0 KB" {
		t.Errorf("expected size_human=1.0 KB, got %v", out["size_human"])
	}
}

func TestNDJSONWriter_MultipleResults(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)
	for i, size := range []int64{512, 2048} {
		r := makeResult(string(rune('0'+i+1)), "/path", size)
		if err := w.Write(r); err != nil {
			t.Fatalf("Write[%d] failed: %v", i, err)
		}
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var out map[string]interface{}
		if err := json.Unmarshal([]byte(line), &out); err != nil {
			t.Errorf("line %d invalid JSON: %v", i, err)
		}
	}
}

func TestJSONWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONWriter(&buf)
	w.Add(makeResult("a", "/path/a", 500))
	w.Add(makeResult("b", "/path/b", 1500000))
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	var out []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Fatalf("invalid JSON array: %v\nbuf: %s", err, buf.String())
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	// Check size_human for 500 bytes
	if out[0]["size_human"] != "500 B" {
		t.Errorf("expected size_human=500 B, got %v", out[0]["size_human"])
	}
	// Check size_human for 1500000 bytes (~1.4 MB)
	if out[1]["size_human"] != "1.4 MB" {
		t.Errorf("expected size_human=1.4 MB, got %v", out[1]["size_human"])
	}
}
