package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleSearchPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8FF85")).Bold(true)
	styleSearchInput  = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#FFFFFF"))
	styleSearchBg     = lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E"))
)

func RenderSearchBar(query string, active bool, width int) string {
	prefix := styleSearchPrefix.Render("Search: ")
	prefixWidth := lipgloss.Width(prefix)

	inputWidth := width - prefixWidth
	if inputWidth < 0 {
		inputWidth = 0
	}

	display := query
	if active {
		display = query + "│"
	}

	// Pad the input area to full width
	dw := len(display)
	if dw < inputWidth {
		display += strings.Repeat(" ", inputWidth-dw)
	}

	input := styleSearchInput.Render(display)
	line := prefix + input
	lw := lipgloss.Width(line)
	if lw < width {
		line += strings.Repeat(" ", width-lw)
	}
	return styleSearchBg.Render(line)
}
