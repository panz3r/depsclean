package tui

const (
	MinWidth     = 80
	MinHeight    = 20
	HeaderHeight = 1
	StatusHeight = 2
)

type Layout struct {
	Width, Height         int
	ListWidth, ListHeight int
	SearchVisible         bool
}

func NewLayout(w, h int, searchVisible bool) Layout {
	l := Layout{
		Width:         w,
		Height:        h,
		SearchVisible: searchVisible,
		ListWidth:     w,
	}
	// header(1) + separator(1) + list + separator(1) + statusbar(2)
	used := HeaderHeight + 1 + 1 + StatusHeight
	if searchVisible {
		used += 2 // search bar + separator
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
