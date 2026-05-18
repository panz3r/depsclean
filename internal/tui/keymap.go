// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tui

type KeyBindings struct {
	Up             []string
	Down           []string
	PageUp         []string
	PageDown       []string
	GotoTop        []string
	GotoBottom     []string
	SearchToggle   []string
	SortCycle      []string
	RowModeToggle  []string
	DetailsToggle  []string
	SelectToggle   []string
	SelectAll      []string
	RangeSelect    []string
	Delete         []string
	DeleteSelected []string
	OpenFolder     []string
	Escape         []string
	Quit           []string
}

func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Up:             []string{"up", "k"},
		Down:           []string{"down", "j"},
		PageUp:         []string{"pgup", "ctrl+b"},
		PageDown:       []string{"pgdown", "ctrl+f"},
		GotoTop:        []string{"home", "g"},
		GotoBottom:     []string{"end", "G"},
		SearchToggle:   []string{"/"},
		SortCycle:      []string{"s"},
		RowModeToggle:  []string{"d"},
		DetailsToggle:  []string{"enter"},
		SelectToggle:   []string{" "},
		SelectAll:      []string{"a"},
		RangeSelect:    []string{"r"},
		Delete:         []string{"x", "delete"},
		DeleteSelected: []string{"X"},
		OpenFolder:     []string{"o"},
		Escape:         []string{"esc"},
		Quit:           []string{"q", "ctrl+c"},
	}
}

func matchKey(key string, bindings []string) bool {
	for _, b := range bindings {
		if key == b {
			return true
		}
	}
	return false
}
