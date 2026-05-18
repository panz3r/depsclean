// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package delete

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/panz3r/depsclean/internal/model"
)

func makeResult(path string) model.Result {
	return model.Result{
		ID:       path,
		Path:     path,
		Basename: filepath.Base(path),
		Status:   model.StatusReady,
	}
}

// TestRefusesRelativePath verifies that Delete rejects paths that are not absolute.
func TestRefusesRelativePath(t *testing.T) {
	d := &Deleter{DryRun: false, Targets: []string{"node_modules"}}
	err := d.Delete(makeResult("relative/node_modules"))
	if err == nil {
		t.Fatal("expected error for relative path, got nil")
	}
}

// TestRefusesDisallowedBasename verifies that Delete rejects a path whose base is not in Targets.
func TestRefusesDisallowedBasename(t *testing.T) {
	d := &Deleter{DryRun: false, Targets: []string{"node_modules"}}
	err := d.Delete(makeResult("/some/project/vendor"))
	if err == nil {
		t.Fatal("expected error for disallowed basename, got nil")
	}
}

// TestRefusesEmptyTargets verifies that Delete refuses to operate when Targets is empty.
func TestRefusesEmptyTargets(t *testing.T) {
	d := &Deleter{DryRun: false, Targets: []string{}}
	err := d.Delete(makeResult("/some/project/node_modules"))
	if err == nil {
		t.Fatal("expected error when Targets is empty, got nil")
	}
}

// TestRefusesNonExistentPath verifies that Delete returns an error for paths that don't exist.
func TestRefusesNonExistentPath(t *testing.T) {
	d := &Deleter{DryRun: false, Targets: []string{"node_modules"}}
	err := d.Delete(makeResult("/nonexistent/node_modules"))
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

// TestRefusesFile verifies that Delete returns an error when the path points to a file (not a dir).
func TestRefusesFile(t *testing.T) {
	// Create a temp file named "node_modules" to test the "not a directory" guard.
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "node_modules")
	if err := os.WriteFile(filePath, []byte("not a dir"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	d := &Deleter{DryRun: false, Targets: []string{"node_modules"}}
	err := d.Delete(makeResult(filePath))
	if err == nil {
		t.Fatal("expected error for file path (not directory), got nil")
	}
}

// TestDryRunDoesNotDelete verifies that DryRun returns nil but leaves the directory intact.
func TestDryRunDoesNotDelete(t *testing.T) {
	tmp := t.TempDir()
	dirPath := filepath.Join(tmp, "node_modules")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	d := &Deleter{DryRun: true, Targets: []string{"node_modules"}}
	if err := d.Delete(makeResult(dirPath)); err != nil {
		t.Fatalf("dry-run returned unexpected error: %v", err)
	}

	// Directory must still exist.
	if _, err := os.Stat(dirPath); err != nil {
		t.Fatalf("directory was deleted during dry-run: %v", err)
	}
}

// TestActualDeletion verifies that Delete recursively removes the target directory.
func TestActualDeletion(t *testing.T) {
	tmp := t.TempDir()
	dirPath := filepath.Join(tmp, "node_modules")
	if err := os.MkdirAll(filepath.Join(dirPath, "subpkg"), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dirPath, "index.js"), []byte(""), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	d := &Deleter{DryRun: false, Targets: []string{"node_modules"}}
	if err := d.Delete(makeResult(dirPath)); err != nil {
		t.Fatalf("Delete returned unexpected error: %v", err)
	}

	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Fatal("expected directory to be removed, but it still exists")
	}
}

// TestDeleteAllGuard verifies that an empty Targets list prevents any deletion.
func TestDeleteAllGuard(t *testing.T) {
	tmp := t.TempDir()
	dirPath := filepath.Join(tmp, "node_modules")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	d := &Deleter{DryRun: false, Targets: nil}
	if err := d.Delete(makeResult(dirPath)); err == nil {
		t.Fatal("expected error with nil Targets, got nil — delete-all guard failed")
	}

	// Directory must still exist.
	if _, err := os.Stat(dirPath); err != nil {
		t.Fatalf("directory was deleted despite guard: %v", err)
	}
}
