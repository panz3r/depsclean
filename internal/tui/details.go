package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/panz3r/depsclean/internal/model"
)

var (
	styleDetailsBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	styleDetailsLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	styleDetailsValue = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	styleDetailsTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8FF85")).Bold(true)
)

func RenderDetailsPanel(r model.Result, width int) string {
	innerWidth := width - 4 // account for border + padding
	if innerWidth < 20 {
		innerWidth = 20
	}

	statusStr := statusLabel(r.Status)

	lines := []string{
		styleDetailsTitle.Render(r.Basename),
		"",
		row("Package Manager", RenderBadge(r.PackageManager), innerWidth),
		row("Status", statusStr, innerWidth),
		row("Path", r.Path, innerWidth),
		row("Project", r.ProjectPath, innerWidth),
	}

	if r.PackageName != "" {
		pkgInfo := r.PackageName
		if r.PackageVersion != "" {
			pkgInfo += "@" + r.PackageVersion
		}
		lines = append(lines, row("Package", pkgInfo, innerWidth))
	}

	lines = append(lines,
		row("Size", formatSize(r.SizeBytes), innerWidth),
		row("Last Modified", formatAge(r.LastModified)+" ("+r.LastModified.Format("2006-01-02")+")", innerWidth),
	)

	if r.ErrorMsg != "" {
		lines = append(lines, row("Error", styleError.Render(r.ErrorMsg), innerWidth))
	}

	content := strings.Join(lines, "\n")
	return styleDetailsBox.Width(innerWidth).Render(content)
}

func row(label, value string, width int) string {
	lbl := styleDetailsLabel.Render(fmt.Sprintf("%-16s", label+":"))
	val := styleDetailsValue.Render(value)
	return lbl + " " + val
}

func statusLabel(s model.Status) string {
	switch s {
	case model.StatusPending:
		return styleSizePending.Render("pending")
	case model.StatusReady:
		return styleSizeReady.Render("ready")
	case model.StatusDeleting:
		return styleSearchPrefix.Render("deleting…")
	case model.StatusDeleted:
		return styleDeleted.Render("deleted")
	case model.StatusError:
		return styleError.Render("error")
	default:
		return "unknown"
	}
}
