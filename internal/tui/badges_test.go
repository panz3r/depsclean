// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tui

import (
	"testing"

	"github.com/panz3r/depsclean/internal/model"
)

func TestRenderBadge(t *testing.T) {
	allManagers := []model.PackageManager{
		model.PackageManagerNPM,
		model.PackageManagerYarn,
		model.PackageManagerPNPM,
		model.PackageManagerBun,
		model.PackageManagerPython,
		model.PackageManagerRust,
		model.PackageManagerGo,
		model.PackageManagerPHP,
		model.PackageManagerRuby,
		model.PackageManagerJava,
		model.PackageManagerUnknown,
	}

	for _, pm := range allManagers {
		t.Run(string(pm), func(t *testing.T) {
			got := RenderBadge(pm)
			if got == "" {
				t.Errorf("RenderBadge(%q) returned empty string", pm)
			}
		})
	}
}
