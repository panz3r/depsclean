package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/panz3r/depsclean/internal/model"
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

	badgePython = lipgloss.NewStyle().
			Background(lipgloss.Color("#3776AB")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	badgeRust = lipgloss.NewStyle().
			Background(lipgloss.Color("#CE422B")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	badgeGo = lipgloss.NewStyle().
			Background(lipgloss.Color("#00ACD7")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	badgePHP = lipgloss.NewStyle().
			Background(lipgloss.Color("#777BB4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	badgeRuby = lipgloss.NewStyle().
			Background(lipgloss.Color("#CC342D")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	badgeJava = lipgloss.NewStyle().
			Background(lipgloss.Color("#ED8B00")).
			Foreground(lipgloss.Color("#FFFFFF")).
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
	case model.PackageManagerPython:
		return badgePython.Render("python")
	case model.PackageManagerRust:
		return badgeRust.Render("rust")
	case model.PackageManagerGo:
		return badgeGo.Render("go")
	case model.PackageManagerPHP:
		return badgePHP.Render("php")
	case model.PackageManagerRuby:
		return badgeRuby.Render("ruby")
	case model.PackageManagerJava:
		return badgeJava.Render("java")
	default:
		return badgeUnknown.Render("?")
	}
}
