package delete

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/panz3r/npclean/internal/model"
)

// Deleter handles safe removal of discovered target directories.
type Deleter struct {
	DryRun  bool
	Targets []string
}

// Delete removes the directory described by result, or simulates removal if DryRun is set.
// Safety checks (in order):
//  1. Path must be absolute.
//  2. Basename must match an allowed target (Targets must not be empty).
//  3. Path must exist and be a directory.
//  4. If DryRun, return nil without touching the filesystem.
//  5. Otherwise, remove the directory recursively.
func (d *Deleter) Delete(result model.Result) error {
	path := result.Path

	// 1. Path must be absolute.
	if !filepath.IsAbs(path) {
		return fmt.Errorf("delete: path must be absolute: %q", path)
	}

	// 2. Basename must match an allowed target; Targets must not be empty.
	base := filepath.Base(path)
	if !d.isAllowedTarget(base) {
		return fmt.Errorf("delete: basename %q is not in the allowed targets list", base)
	}

	// 3. Path must exist and be a directory.
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("delete: path does not exist: %q", path)
		}
		return fmt.Errorf("delete: stat %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("delete: path is not a directory: %q", path)
	}

	// 4. Dry-run: skip actual removal.
	if d.DryRun {
		return nil
	}

	// 5. Remove the directory recursively.
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("delete: remove %q: %w", path, err)
	}
	return nil
}

// isAllowedTarget reports whether base is in d.Targets.
// Returns false if Targets is empty (guards against accidental delete-all).
func (d *Deleter) isAllowedTarget(base string) bool {
	if len(d.Targets) == 0 {
		return false
	}
	for _, t := range d.Targets {
		if t == base {
			return true
		}
	}
	return false
}
