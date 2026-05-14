package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/panz3r/npclean/internal/tui"
)

// RunTUI launches the interactive TUI.
func RunTUI() error {
	m := tui.New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
