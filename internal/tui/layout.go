// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tui

const (
	MinWidth      = 80
	MinHeight     = 20
	HeaderHeight  = 3
	StatusHeight  = 2
	DetailsHeight = 12
)

type Layout struct {
	Width, Height         int
	ListWidth, ListHeight int
	SearchVisible         bool
}

func NewLayout(w, h int, searchVisible bool, detailsVisible bool) Layout {
	l := Layout{
		Width:         w,
		Height:        h,
		SearchVisible: searchVisible,
		ListWidth:     w,
	}
	// header(3) + separator(1) + list + separator(1) + statusbar(2)
	used := HeaderHeight + 1 + 1 + StatusHeight
	if searchVisible {
		used += 2 // search bar + separator
	}
	if detailsVisible {
		used += DetailsHeight + 1 // details panel + separator
	}
	l.ListHeight = h - used
	if l.ListHeight < 0 {
		l.ListHeight = 0
	}
	return l
}

func (l Layout) TooSmall() bool {
	return l.Width < MinWidth || l.Height < MinHeight
}
