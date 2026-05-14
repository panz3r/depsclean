package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/panz3r/npclean/internal/model"
)

var (
	badgeNpm = lipgloss.NewStyle().
		Background(lipgloss.Color("#CB3837")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	badgeYarn = lipgloss.NewStyle().
		Background(lipgloss.Color("#2C8EBB")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	badgePnpm = lipgloss.NewStyle().
		Background(lipgloss.Color("#F69220")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1)

	badgeBun = lipgloss.NewStyle().
		Background(lipgloss.Color("#FBF0DF")).
		Foreground(lipgloss.Color("#1A1A1A")).
		Bold(true).
		Padding(0, 1)

	badgeUnknown = lipgloss.NewStyle().
		Background(lipgloss.Color("#555555")).
		Foreground(lipgloss.Color("#CCCCCC")).
		Bold(true).
		Padding(0, 1)
)

func RenderBadge(pm model.PackageManager) string {
	switch pm {
	case model.PackageManagerNPM:
		return badgeNpm.Render("npm")
	case model.PackageManagerYarn:
		return badgeYarn.Render("yarn")
	case model.PackageManagerPNPM:
		return badgePnpm.Render("pnpm")
	case model.PackageManagerBun:
		return badgeBun.Render("bun")
	default:
		return badgeUnknown.Render("?")
	}
}
