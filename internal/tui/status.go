package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleStatusBg   = lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E"))
	styleStatusKey  = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8FF85")).Bold(true)
	styleStatusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	styleStatusDone = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88"))
	styleStatusScan = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
)

func RenderStatusBar(total, visible, selected int, sortLabel, rowModeLabel string, scanDone bool, width int) string {
	// Line 1: key hints
	hints := []string{
		styleStatusKey.Render("↑↓") + styleStatusText.Render(" navigate"),
		styleStatusKey.Render("/") + styleStatusText.Render(" search"),
		styleStatusKey.Render("s") + styleStatusText.Render(" sort:") + styleStatusText.Render(sortLabel),
		styleStatusKey.Render("d") + styleStatusText.Render(":") + styleStatusText.Render(rowModeLabel),
		styleStatusKey.Render("enter") + styleStatusText.Render(" details"),
		styleStatusKey.Render("space") + styleStatusText.Render(" select"),
		styleStatusKey.Render("q") + styleStatusText.Render(" quit"),
	}
	line1 := " " + strings.Join(hints, styleStatusText.Render("  "))
	lw := lipgloss.Width(line1)
	if lw < width {
		line1 += strings.Repeat(" ", width-lw)
	}

	// Line 2: stats
	var scanStatus string
	if scanDone {
		scanStatus = styleStatusDone.Render("✓ scan complete")
	} else {
		scanStatus = styleStatusScan.Render("⠿ scanning…")
	}

	stats := fmt.Sprintf(" %s  %s  %s  %s",
		styleStatusText.Render(fmt.Sprintf("%d results", total)),
		styleStatusText.Render(fmt.Sprintf("%d visible", visible)),
		styleStatusText.Render(fmt.Sprintf("%d selected", selected)),
		scanStatus,
	)
	sw := lipgloss.Width(stats)
	if sw < width {
		stats += strings.Repeat(" ", width-sw)
	}

	line1Styled := styleStatusBg.Width(width).Render(line1)
	line2Styled := styleStatusBg.Width(width).Render(stats)

	return line1Styled + "\n" + line2Styled
}
