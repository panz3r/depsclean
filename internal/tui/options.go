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
