// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package scan

import (
	"path/filepath"
	"testing"
)

func TestIsSensitivePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"unix /etc exact", "/etc", true},
		{"unix /usr exact", "/usr", true},
		{"unix /usr/bin sub-path", "/usr/bin", true},
		{"unix /var/log sub-path", "/var/log", true},
		{"macOS /Applications exact", "/Applications", true},
		{"macOS /System/Library sub-path", "/System/Library", true},
		{"macOS /private/tmp sub-path", "/private/tmp", true},
		{"non-sensitive /home", "/home", false},
		{"non-sensitive /projects", "/projects", false},
		{"empty string", "", false},
		{"/usr/local/bin triple-nested", "/usr/local/bin", true},
		{"not prefix /usrextra", "/usrextra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSensitivePath(tt.path)
			if got != tt.want {
				t.Errorf("IsSensitivePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsHidden(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"single dot", ".", true},
		{"dot git", ".git", true},
		{"dot env", ".env", true},
		{"node_modules not hidden", "node_modules", false},
		{"empty string", "", false},
		{"normal dir", "src", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHidden(tt.input)
			if got != tt.want {
				t.Errorf("IsHidden(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPathDepth(t *testing.T) {
	root := "/tmp/testroot"

	tests := []struct {
		name string
		p    string
		root string
		want int
	}{
		{"same as root", root, root, 0},
		{"one level deep", filepath.Join(root, "a"), root, 1},
		{"two levels deep", filepath.Join(root, "a", "b"), root, 2},
		{"three levels deep", filepath.Join(root, "a", "b", "c"), root, 3},
		{"path not under root", "/other/path", root, -1},
		{"parent of root", "/tmp", root, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathDepth(tt.p, tt.root)
			if got != tt.want {
				t.Errorf("PathDepth(%q, %q) = %v, want %v", tt.p, tt.root, got, tt.want)
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{"match full path", "/project/vendor", []string{"/project/vendor"}, true},
		{"match basename", "/project/vendor", []string{"vendor"}, true},
		{"glob basename", "/project/dist", []string{"d*"}, true},
		{"no match", "/project/src", []string{"vendor", "dist"}, false},
		{"empty patterns", "/project/vendor", []string{}, false},
		{"glob full path", "/tmp/node_modules", []string{"/tmp/*"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExcluded(tt.path, tt.patterns)
			if got != tt.want {
				t.Errorf("IsExcluded(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}
