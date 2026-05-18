// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import "sort"

// Profile defines a named set of scan targets with metadata.
type Profile struct {
	Name        string
	Description string
	Targets     []string
}

// BuiltinProfiles is the registry of built-in profiles.
var BuiltinProfiles = map[string]Profile{
	"node": {
		Name:        "node",
		Description: "Node.js – removes node_modules directories",
		Targets:     []string{"node_modules"},
	},
	"python": {
		Name:        "python",
		Description: "Python virtual envs and build artefacts",
		Targets:     []string{".venv", "venv", "__pycache__", ".pytest_cache", ".mypy_cache", "dist", "build", "*.egg-info"},
	},
	"rust": {
		Name:        "rust",
		Description: "Rust build output",
		Targets:     []string{"target"},
	},
	"go": {
		Name:        "go",
		Description: "Go vendor directories",
		Targets:     []string{"vendor"},
	},
	"php": {
		Name:        "php",
		Description: "Composer vendor",
		Targets:     []string{"vendor"},
	},
	"ruby": {
		Name:        "ruby",
		Description: "Bundler vendor",
		Targets:     []string{"vendor/bundle", ".bundle"},
	},
	"java": {
		Name:        "java",
		Description: "Maven/Gradle build output",
		Targets:     []string{"target", ".gradle", "build"},
	},
}

func init() {
	// Build "all" profile as the union of all other profiles' targets.
	seen := make(map[string]bool)
	var allTargets []string
	names := make([]string, 0, len(BuiltinProfiles))
	for name := range BuiltinProfiles {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for _, t := range BuiltinProfiles[name].Targets {
			if !seen[t] {
				seen[t] = true
				allTargets = append(allTargets, t)
			}
		}
	}
	BuiltinProfiles["all"] = Profile{
		Name:        "all",
		Description: "All targets from every built-in profile",
		Targets:     allTargets,
	}
}

// LookupProfile returns the Profile for name, and a bool indicating if it was found.
func LookupProfile(name string) (Profile, bool) {
	p, ok := BuiltinProfiles[name]
	return p, ok
}

// ListProfiles returns all built-in profile names in sorted order.
func ListProfiles() []string {
	names := make([]string, 0, len(BuiltinProfiles))
	for name := range BuiltinProfiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
