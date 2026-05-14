package tui

type KeyBindings struct {
	Up            []string
	Down          []string
	PageUp        []string
	PageDown      []string
	GotoTop       []string
	GotoBottom    []string
	SearchToggle  []string
	SortCycle     []string
	RowModeToggle []string
	DetailsToggle []string
	SelectToggle  []string
	Escape        []string
	Quit          []string
}

func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Up:            []string{"up", "k"},
		Down:          []string{"down", "j"},
		PageUp:        []string{"pgup", "ctrl+b"},
		PageDown:      []string{"pgdown", "ctrl+f"},
		GotoTop:       []string{"home", "g"},
		GotoBottom:    []string{"end", "G"},
		SearchToggle:  []string{"/"},
		SortCycle:     []string{"s"},
		RowModeToggle: []string{"d"},
		DetailsToggle: []string{"enter"},
		SelectToggle:  []string{" "},
		Escape:        []string{"esc"},
		Quit:          []string{"q", "ctrl+c"},
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
