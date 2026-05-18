// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package scan

import (
	"path/filepath"
	"strings"
)

// sensitivePaths is the set of known sensitive system directory paths.
var sensitivePaths = []string{
	// Unix
	"/bin", "/boot", "/dev", "/etc", "/lib", "/lib64",
	"/proc", "/root", "/run", "/sbin", "/srv", "/sys", "/tmp", "/usr", "/var",
	// macOS
	"/Applications", "/Library", "/Network", "/System", "/Volumes", "/private",
	// Windows
	`C:\Windows`, `C:\Program Files`, `C:\Program Files (x86)`,
}

// IsSensitivePath reports whether path is or is inside a known sensitive system directory.
func IsSensitivePath(path string) bool {
	clean := filepath.Clean(path)
	for _, s := range sensitivePaths {
		if clean == s {
			return true
		}
		if strings.HasPrefix(clean, s+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// IsHidden reports whether name starts with '.'.
func IsHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// PathDepth returns the number of path segments in p relative to root.
// Returns 0 if p == root, -1 on error or if p is not under root.
func PathDepth(p, root string) int {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return -1
	}
	if rel == "." {
		return 0
	}
	if strings.HasPrefix(rel, "..") {
		return -1
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

// IsExcluded reports whether path matches any of the provided glob patterns
// (matched against both the full path and the basename).
func IsExcluded(path string, patterns []string) bool {
	base := filepath.Base(path)
	for _, pat := range patterns {
		if matched, _ := filepath.Match(pat, path); matched {
			return true
		}
		if matched, _ := filepath.Match(pat, base); matched {
			return true
		}
	}
	return false
}
