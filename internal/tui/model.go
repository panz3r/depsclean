package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/panz3r/npclean/internal/model"
)

// Messages
type ResultMsg struct{ R model.Result }
type ScanDoneMsg struct{ Total int }

type ScanState int

const (
	ScanStateScanning ScanState = iota
	ScanStateDone
)

type Model struct {
	keys           KeyBindings
	layout         Layout
	allResults     []model.Result
	visibleResults []model.Result
	selected       map[string]bool
	cursor         int
	offset         int
	searchQuery    string
	searchMode     bool
	sortMode       SortMode
	rowMode        RowMode
	showDetails    bool
	scanState      ScanState
	scanFound      int
	errors         []string
}

func New() Model {
	return Model{
		keys:     DefaultKeyBindings(),
		layout:   NewLayout(80, 24, false),
		selected: make(map[string]bool),
		sortMode: SortBySizeDesc,
		rowMode:  RowModeCompact,
	}
}

func (m Model) Init() tea.Cmd {
	fixtures := FixtureResults()
	cmds := make([]tea.Cmd, 0, len(fixtures)+1)
	for _, r := range fixtures {
		r := r
		cmds = append(cmds, func() tea.Msg { return ResultMsg{R: r} })
	}
	cmds = append(cmds, func() tea.Msg { return ScanDoneMsg{Total: len(fixtures)} })
	return tea.Sequence(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = NewLayout(msg.Width, msg.Height, m.searchMode)

	case ResultMsg:
		m.allResults = append(m.allResults, msg.R)
		m.scanFound++
		m.refilterAndSort()

	case ScanDoneMsg:
		m.scanState = ScanStateDone

	case tea.KeyMsg:
		if m.searchMode {
			return m.updateSearch(msg)
		}
		return m.updateNormal(msg)
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.layout = NewLayout(m.layout.Width, m.layout.Height, false)
		m.refilterAndSort()
	case "enter":
		m.searchMode = false
		m.layout = NewLayout(m.layout.Width, m.layout.Height, false)
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.refilterAndSort()
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.refilterAndSort()
		}
	}
	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	linesPerRow := 1
	if m.rowMode == RowModeDetails {
		linesPerRow = 2
	}
	pageSize := m.layout.ListHeight / linesPerRow
	if pageSize < 1 {
		pageSize = 1
	}

	switch {
	case matchKey(key, m.keys.Quit):
		return m, tea.Quit

	case matchKey(key, m.keys.Up):
		m.moveCursor(-1)

	case matchKey(key, m.keys.Down):
		m.moveCursor(1)

	case matchKey(key, m.keys.PageUp):
		m.moveCursor(-pageSize)

	case matchKey(key, m.keys.PageDown):
		m.moveCursor(pageSize)

	case matchKey(key, m.keys.GotoTop):
		m.cursor = 0
		m.offset = 0

	case matchKey(key, m.keys.GotoBottom):
		if len(m.visibleResults) > 0 {
			m.cursor = len(m.visibleResults) - 1
		}
		m.clampOffset()

	case matchKey(key, m.keys.SearchToggle):
		m.searchMode = true
		m.layout = NewLayout(m.layout.Width, m.layout.Height, true)

	case matchKey(key, m.keys.SortCycle):
		m.sortMode = NextSortMode(m.sortMode)
		m.refilterAndSort()

	case matchKey(key, m.keys.RowModeToggle):
		if m.rowMode == RowModeCompact {
			m.rowMode = RowModeDetails
		} else {
			m.rowMode = RowModeCompact
		}

	case matchKey(key, m.keys.DetailsToggle):
		m.showDetails = !m.showDetails

	case matchKey(key, m.keys.SelectToggle):
		if len(m.visibleResults) > 0 && m.cursor < len(m.visibleResults) {
			id := m.visibleResults[m.cursor].ID
			m.selected[id] = !m.selected[id]
			if !m.selected[id] {
				delete(m.selected, id)
			}
		}

	case matchKey(key, m.keys.Escape):
		m.selected = make(map[string]bool)
		m.showDetails = false
	}

	return m, nil
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if len(m.visibleResults) > 0 && m.cursor >= len(m.visibleResults) {
		m.cursor = len(m.visibleResults) - 1
	}
	m.clampOffset()
}

func (m *Model) clampOffset() {
	linesPerRow := 1
	if m.rowMode == RowModeDetails {
		linesPerRow = 2
	}
	capacity := m.layout.ListHeight / linesPerRow
	if capacity < 1 {
		capacity = 1
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+capacity {
		m.offset = m.cursor - capacity + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) refilterAndSort() {
	q := strings.ToLower(m.searchQuery)
	visible := make([]model.Result, 0, len(m.allResults))
	for _, r := range m.allResults {
		if q == "" {
			visible = append(visible, r)
			continue
		}
		haystack := strings.ToLower(r.ProjectPath + " " + r.PackageName + " " + r.PackageVersion + " " + r.Basename)
		if strings.Contains(haystack, q) {
			visible = append(visible, r)
		}
	}

	sort.Slice(visible, func(i, j int) bool {
		a, b := visible[i], visible[j]
		switch m.sortMode {
		case SortBySizeDesc:
			return a.SizeBytes > b.SizeBytes
		case SortBySizeAsc:
			return a.SizeBytes < b.SizeBytes
		case SortByNameAsc:
			na := a.PackageName
			if na == "" {
				na = a.Basename
			}
			nb := b.PackageName
			if nb == "" {
				nb = b.Basename
			}
			return na < nb
		case SortByPathAsc:
			return a.ProjectPath < b.ProjectPath
		case SortByNewest:
			return a.LastModified.After(b.LastModified)
		case SortByOldest:
			return a.LastModified.Before(b.LastModified)
		default:
			return a.SizeBytes > b.SizeBytes
		}
	})

	m.visibleResults = visible

	// Clamp cursor
	if len(m.visibleResults) == 0 {
		m.cursor = 0
		m.offset = 0
	} else if m.cursor >= len(m.visibleResults) {
		m.cursor = len(m.visibleResults) - 1
	}
	m.clampOffset()
}

var (
	styleHeader      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E8FF85"))
	styleHeaderRight = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	styleSeparator   = lipgloss.NewStyle().Foreground(lipgloss.Color("#333366"))
	styleTooSmall    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F57")).Bold(true)
)

func (m Model) View() string {
	if m.layout.TooSmall() {
		return styleTooSmall.Render("Terminal too small (min 80×20)")
	}

	var sb strings.Builder
	w := m.layout.Width

	// Header
	title := styleHeader.Render("npclean")
	var rightText string
	if m.scanState == ScanStateDone {
		rightText = styleStatusDone.Render("✓ scan complete") + styleHeaderRight.Render(fmt.Sprintf(" (%d found)", m.scanFound))
	} else {
		rightText = styleStatusScan.Render("scanning…") + styleHeaderRight.Render(fmt.Sprintf(" (%d found)", m.scanFound))
	}
	titleWidth := lipgloss.Width(title)
	rightWidth := lipgloss.Width(rightText)
	gap := w - titleWidth - rightWidth
	if gap < 1 {
		gap = 1
	}
	header := title + strings.Repeat(" ", gap) + rightText
	sb.WriteString(header)
	sb.WriteString("\n")

	// Separator
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", w)))
	sb.WriteString("\n")

	// Search bar
	if m.searchMode {
		sb.WriteString(RenderSearchBar(m.searchQuery, true, w))
		sb.WriteString("\n")
		sb.WriteString(styleSeparator.Render(strings.Repeat("─", w)))
		sb.WriteString("\n")
	} else if m.searchQuery != "" {
		sb.WriteString(RenderSearchBar(m.searchQuery, false, w))
		sb.WriteString("\n")
		sb.WriteString(styleSeparator.Render(strings.Repeat("─", w)))
		sb.WriteString("\n")
	}

	// List
	linesPerRow := 1
	if m.rowMode == RowModeDetails {
		linesPerRow = 2
	}
	capacity := m.layout.ListHeight / linesPerRow
	if capacity < 1 {
		capacity = 1
	}

	rendered := 0
	for i := m.offset; i < len(m.visibleResults) && rendered < capacity; i++ {
		r := m.visibleResults[i]
		isCursor := i == m.cursor
		isSelected := m.selected[r.ID]
		var rowStr string
		if m.rowMode == RowModeDetails {
			rowStr = RenderDetailsRow(r, isCursor, isSelected, w)
		} else {
			rowStr = RenderCompactRow(r, isCursor, isSelected, w)
		}
		sb.WriteString(rowStr)
		sb.WriteString("\n")
		rendered++
	}

	// Fill remaining list space
	fillerLines := m.layout.ListHeight - rendered*linesPerRow
	for i := 0; i < fillerLines; i++ {
		sb.WriteString(strings.Repeat(" ", w))
		sb.WriteString("\n")
	}

	// Separator
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", w)))
	sb.WriteString("\n")

	// Status bar
	rowModeLabel := "compact"
	if m.rowMode == RowModeDetails {
		rowModeLabel = "details"
	}
	sb.WriteString(RenderStatusBar(
		len(m.allResults),
		len(m.visibleResults),
		len(m.selected),
		SortModeLabel(m.sortMode),
		rowModeLabel,
		m.scanState == ScanStateDone,
		w,
	))

	return sb.String()
}
