// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tui

type SortMode int

const (
	SortBySizeDesc SortMode = iota
	SortBySizeAsc
	SortByNameAsc
	SortByPathAsc
	SortByNewest
	SortByOldest
	sortModeCount
)

func SortModeLabel(m SortMode) string {
	switch m {
	case SortBySizeDesc:
		return "size↓"
	case SortBySizeAsc:
		return "size↑"
	case SortByNameAsc:
		return "name↑"
	case SortByPathAsc:
		return "path↑"
	case SortByNewest:
		return "newest"
	case SortByOldest:
		return "oldest"
	default:
		return "size↓"
	}
}

func NextSortMode(m SortMode) SortMode {
	return (m + 1) % sortModeCount
}

type RowMode int

const (
	RowModeCompact RowMode = iota
	RowModeDetails
)
