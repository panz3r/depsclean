package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/panz3r/depsclean/internal/model"
)

var (
	styleCursorBg    = lipgloss.NewStyle().Background(lipgloss.Color("#5A3EC8")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	styleSelectedBg  = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FFFFFF"))
	styleNormalPath  = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	styleProjectName = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	styleSizeReady   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8FF85"))
	styleSizePending = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	styleDeleted     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Strikethrough(true)
	styleDeleting    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F57"))
	styleMeta        = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	styleCheckmark   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)

func formatSize(bytes int64) string {
	const (
		GB = 1 << 30
		MB = 1 << 20
		KB = 1 << 10
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
	}
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

func RenderCompactRow(r model.Result, isCursor bool, isSelected bool, width int) string {
	badge := RenderBadge(r.PackageManager)
	badgeWidth := lipgloss.Width(badge)

	selectMark := "  "
	if isSelected {
		selectMark = styleCheckmark.Render("✓ ")
	}

	var sizeStr string
	sizeWidth := 9
	if r.Status == model.StatusPending {
		sizeStr = styleSizePending.Render(fmt.Sprintf("%*s", sizeWidth, "…"))
	} else if r.Status == model.StatusDeleted {
		sizeStr = styleDeleted.Render(fmt.Sprintf("%*s", sizeWidth, formatSize(r.SizeBytes)))
	} else if r.Status == model.StatusDeleting {
		sizeStr = styleDeleting.Render(fmt.Sprintf("%*s", sizeWidth, "deleting…"))
	} else if r.Status == model.StatusError {
		sizeStr = styleError.Render(fmt.Sprintf("%*s", sizeWidth, "ERROR"))
	} else {
		sizeStr = styleSizeReady.Render(fmt.Sprintf("%*s", sizeWidth, formatSize(r.SizeBytes)))
	}
	sizeRenderWidth := lipgloss.Width(sizeStr)

	// Available width for name + path
	// 2 (selectMark) + badgeWidth + 1 (space) + namePathWidth + 1 (space) + sizeRenderWidth
	namePathWidth := width - 2 - badgeWidth - 1 - 1 - sizeRenderWidth
	if namePathWidth < 10 {
		namePathWidth = 10
	}

	name := r.Basename
	if r.PackageName != "" {
		name = r.PackageName
	}

	// Show parent path truncated
	parentPath := r.ProjectPath
	nameLen := len(name)
	pathAvail := namePathWidth - nameLen - 1
	if pathAvail < 0 {
		pathAvail = 0
	}
	truncated := truncatePath(parentPath, pathAvail)

	var namePart, pathPart string
	if r.Status == model.StatusDeleted {
		namePart = styleDeleted.Render(name)
		pathPart = styleDeleted.Render(truncated)
	} else if r.Status == model.StatusDeleting {
		namePart = styleDeleting.Render(name)
		pathPart = styleDeleting.Render(truncated)
	} else if r.Status == model.StatusError {
		namePart = styleError.Render(name)
		pathPart = styleError.Render(truncated)
	} else {
		namePart = styleProjectName.Render(name)
		pathPart = styleNormalPath.Render(truncated)
	}

	namePath := namePart + " " + pathPart
	// Pad to fill available space
	npWidth := lipgloss.Width(namePath)
	if npWidth < namePathWidth {
		namePath += strings.Repeat(" ", namePathWidth-npWidth)
	}

	line := selectMark + badge + " " + namePath + " " + sizeStr

	if isCursor {
		line = styleCursorBg.Width(width).Render(line)
	} else if isSelected {
		line = styleSelectedBg.Width(width).Render(line)
	} else {
		// Pad to full width
		lw := lipgloss.Width(line)
		if lw < width {
			line += strings.Repeat(" ", width-lw)
		}
	}

	return line
}

func RenderDetailsRow(r model.Result, isCursor bool, isSelected bool, width int) string {
	line1 := RenderCompactRow(r, isCursor, isSelected, width)

	// Line 2: metadata
	indent := "    "
	var meta string
	if r.PackageName != "" && r.PackageVersion != "" {
		meta = r.PackageName + "@" + r.PackageVersion
	} else if r.PackageName != "" {
		meta = r.PackageName
	}
	if !r.LastModified.IsZero() {
		if meta != "" {
			meta += "  ·  "
		}
		meta += "modified " + formatAge(r.LastModified)
	}
	if r.Status == model.StatusError && r.ErrorMsg != "" {
		if meta != "" {
			meta += "  ·  "
		}
		meta += styleError.Render(r.ErrorMsg)
	}

	metaLine := indent + styleMeta.Render(meta)
	metaWidth := lipgloss.Width(metaLine)
	if metaWidth < width {
		metaLine += strings.Repeat(" ", width-metaWidth)
	}

	if isCursor {
		metaLine = styleCursorBg.Width(width).Render(indent + styleMeta.Render(meta))
	} else if isSelected {
		metaLine = styleSelectedBg.Width(width).Render(indent + styleMeta.Render(meta))
	}

	return line1 + "\n" + metaLine
}
