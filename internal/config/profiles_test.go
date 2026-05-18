// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"sort"
	"testing"
)

func TestLookupProfile_Node(t *testing.T) {
	p, ok := LookupProfile("node")
	if !ok {
		t.Fatal("expected node profile to exist")
	}
	if len(p.Targets) == 0 || p.Targets[0] != "node_modules" {
		t.Errorf("unexpected node targets: %v", p.Targets)
	}
}

func TestLookupProfile_NotFound(t *testing.T) {
	_, ok := LookupProfile("nonexistent")
	if ok {
		t.Error("expected nonexistent profile to not be found")
	}
}

func TestListProfiles_Sorted(t *testing.T) {
	names := ListProfiles()
	if len(names) == 0 {
		t.Fatal("expected at least one profile")
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("profiles not sorted: %v", names)
	}
	// Check that all built-in profiles are present
	for name := range BuiltinProfiles {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("profile %q not in ListProfiles result", name)
		}
	}
}

func TestAllProfile_UnionOfAll(t *testing.T) {
	all, ok := LookupProfile("all")
	if !ok {
		t.Fatal("expected all profile to exist")
	}
	// Verify it contains targets from other profiles
	allSet := make(map[string]bool)
	for _, target := range all.Targets {
		allSet[target] = true
	}
	for name, p := range BuiltinProfiles {
		if name == "all" {
			continue
		}
		for _, target := range p.Targets {
			if !allSet[target] {
				t.Errorf("all profile missing target %q from profile %q", target, name)
			}
		}
	}
}
