package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	deletepkg "github.com/panz3r/npclean/internal/delete"
	"github.com/panz3r/npclean/internal/model"
	"github.com/panz3r/npclean/internal/platform"
	"github.com/panz3r/npclean/internal/scan"
)

// Messages
type ResultMsg struct{ R model.Result }
type ScanDoneMsg struct{ Total int }

type deleteResultMsg struct {
	id  string
	err error
}

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
	startCmd       tea.Cmd
	resultsIndex   map[string]int     // ID → index in allResults (for O(1) in-place update)
	cursorID       string             // stable cursor identity across async updates and re-sorts
	deleter        *deletepkg.Deleter // nil in demo mode
	rangeAnchor    string             // ID of range selection anchor; "" = not active
	dryRun         bool
}

func New() Model {
	return newModel(FixtureStartCmd())
}

func NewWithScan(startCmd tea.Cmd) Model {
	return newModel(startCmd)
}

func NewWithScanAndDeleter(startCmd tea.Cmd, d *deletepkg.Deleter) Model {
	m := newModel(startCmd)
	m.deleter = d
	if d != nil {
		m.dryRun = d.DryRun
	}
	return m
}

func newModel(startCmd tea.Cmd) Model {
	return Model{
		keys:         DefaultKeyBindings(),
		layout:       NewLayout(80, 24, false, false),
		selected:     make(map[string]bool),
		resultsIndex: make(map[string]int),
		sortMode:     SortBySizeDesc,
		rowMode:      RowModeCompact,
		startCmd:     startCmd,
	}
}

func (m Model) Init() tea.Cmd {
	return m.startCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = NewLayout(msg.Width, msg.Height, m.searchMode, m.showDetails)

	case scanEventMsg:
		cmd := WaitForScanEvent(msg.ch) // schedule next read
		switch e := msg.ev.(type) {
		case scan.DiscoveredEvent:
			m.addResult(e.Result)
		case scan.AnalyzedEvent:
			m.updateResult(e.Result)
		case scan.ErrorEvent:
			m.errors = append(m.errors, e.Err.Error())
		case scan.DoneEvent:
			m.scanState = ScanStateDone
			m.scanFound = e.Total
		}
		return m, cmd

	case ResultMsg:
		m.addResult(msg.R)

	case ScanDoneMsg:
		m.scanState = ScanStateDone

	case deleteResultMsg:
		m.applyDeleteResult(msg)

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
		m.layout = NewLayout(m.layout.Width, m.layout.Height, false, m.showDetails)
		m.refilterAndSort()
	case "enter":
		m.searchMode = false
		m.layout = NewLayout(m.layout.Width, m.layout.Height, false, m.showDetails)
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
		if len(m.visibleResults) > 0 {
			m.cursorID = m.visibleResults[0].ID
		}

	case matchKey(key, m.keys.GotoBottom):
		if len(m.visibleResults) > 0 {
			m.cursor = len(m.visibleResults) - 1
		}
		m.clampOffset()
		if len(m.visibleResults) > 0 {
			m.cursorID = m.visibleResults[m.cursor].ID
		}

	case matchKey(key, m.keys.SearchToggle):
		m.searchMode = true
		m.layout = NewLayout(m.layout.Width, m.layout.Height, true, m.showDetails)

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
		m.layout = NewLayout(m.layout.Width, m.layout.Height, m.searchMode, m.showDetails)

	case matchKey(key, m.keys.SelectToggle):
		if len(m.visibleResults) > 0 && m.cursor < len(m.visibleResults) {
			id := m.visibleResults[m.cursor].ID
			m.selected[id] = !m.selected[id]
			if !m.selected[id] {
				delete(m.selected, id)
			}
		}

	case matchKey(key, m.keys.SelectAll):
		m.toggleSelectAll()

	case matchKey(key, m.keys.RangeSelect):
		m.handleRangeSelect()

	case matchKey(key, m.keys.Delete):
		if m.deleter != nil && len(m.visibleResults) > 0 && m.cursor < len(m.visibleResults) {
			r := m.visibleResults[m.cursor]
			if r.Status != model.StatusDeleting && r.Status != model.StatusDeleted {
				m.setStatus(r.ID, model.StatusDeleting)
				return m, m.makeDeleteCmd(r)
			}
		}

	case matchKey(key, m.keys.DeleteSelected):
		return m, m.deleteSelected()

	case matchKey(key, m.keys.OpenFolder):
		if len(m.visibleResults) > 0 && m.cursor < len(m.visibleResults) {
			r := m.visibleResults[m.cursor]
			if err := platform.Open(r.ProjectPath); err != nil {
				m.errors = append(m.errors, "open folder: "+err.Error())
			}
		}

	case matchKey(key, m.keys.Escape):
		m.selected = make(map[string]bool)
		m.showDetails = false
		m.rangeAnchor = ""
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
	if m.cursor >= 0 && m.cursor < len(m.visibleResults) {
		m.cursorID = m.visibleResults[m.cursor].ID
	}
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

func (m *Model) addResult(r model.Result) {
	if _, exists := m.resultsIndex[r.ID]; exists {
		return // already known
	}
	m.resultsIndex[r.ID] = len(m.allResults)
	m.allResults = append(m.allResults, r)
	m.scanFound++
	m.refilterAndSort()
}

func (m *Model) updateResult(r model.Result) {
	idx, ok := m.resultsIndex[r.ID]
	if !ok {
		// Not yet in the list - treat as new
		m.addResult(r)
		return
	}
	m.allResults[idx] = r
	m.refilterAndSort()
}

// makeDeleteCmd returns a Cmd that runs deletion in the background.
func (m *Model) makeDeleteCmd(r model.Result) tea.Cmd {
	d := m.deleter
	return func() tea.Msg {
		err := d.Delete(r)
		return deleteResultMsg{id: r.ID, err: err}
	}
}

// applyDeleteResult processes the result of an async deletion.
func (m *Model) applyDeleteResult(msg deleteResultMsg) {
	idx, ok := m.resultsIndex[msg.id]
	if !ok {
		return
	}
	if msg.err != nil {
		m.allResults[idx].Status = model.StatusError
		m.allResults[idx].ErrorMsg = msg.err.Error()
	} else {
		m.allResults[idx].Status = model.StatusDeleted
		m.allResults[idx].ErrorMsg = ""
	}
	m.refilterAndSort()
}

// setStatus updates the status of the result with the given ID in allResults.
func (m *Model) setStatus(id string, s model.Status) {
	idx, ok := m.resultsIndex[id]
	if !ok {
		return
	}
	m.allResults[idx].Status = s
	m.refilterAndSort()
}

// toggleSelectAll selects all visible non-deleted rows, or deselects all if all are already selected.
func (m *Model) toggleSelectAll() {
	eligible := make([]string, 0)
	for _, r := range m.visibleResults {
		if r.Status != model.StatusDeleted && r.Status != model.StatusDeleting {
			eligible = append(eligible, r.ID)
		}
	}
	allSelected := len(eligible) > 0
	for _, id := range eligible {
		if !m.selected[id] {
			allSelected = false
			break
		}
	}
	if allSelected {
		m.selected = make(map[string]bool)
	} else {
		for _, id := range eligible {
			m.selected[id] = true
		}
	}
}

// handleRangeSelect sets the range anchor on first press, then selects the range on second press.
func (m *Model) handleRangeSelect() {
	if len(m.visibleResults) == 0 || m.cursor >= len(m.visibleResults) {
		return
	}
	cursorID := m.visibleResults[m.cursor].ID
	if m.rangeAnchor == "" {
		m.rangeAnchor = cursorID
		return
	}
	// Find anchor position in visibleResults.
	anchorIdx := -1
	for i, r := range m.visibleResults {
		if r.ID == m.rangeAnchor {
			anchorIdx = i
			break
		}
	}
	if anchorIdx == -1 {
		// Anchor is no longer visible; reset and start fresh.
		m.rangeAnchor = cursorID
		return
	}
	lo, hi := anchorIdx, m.cursor
	if lo > hi {
		lo, hi = hi, lo
	}
	for i := lo; i <= hi; i++ {
		r := m.visibleResults[i]
		if r.Status != model.StatusDeleted && r.Status != model.StatusDeleting {
			m.selected[r.ID] = true
		}
	}
	m.rangeAnchor = ""
}

// deleteSelected initiates deletion for all selected items.
func (m *Model) deleteSelected() tea.Cmd {
	if m.deleter == nil {
		return nil
	}
	cmds := make([]tea.Cmd, 0)
	for id := range m.selected {
		idx, ok := m.resultsIndex[id]
		if !ok {
			continue
		}
		r := m.allResults[idx]
		if r.Status == model.StatusDeleting || r.Status == model.StatusDeleted {
			continue
		}
		m.setStatus(id, model.StatusDeleting)
		r = m.allResults[idx] // re-fetch after setStatus
		cmds = append(cmds, m.makeDeleteCmd(r))
	}
	m.selected = make(map[string]bool)
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m *Model) refilterAndSort() {
	// Save cursor ID before rebuild
	if m.cursor >= 0 && m.cursor < len(m.visibleResults) {
		m.cursorID = m.visibleResults[m.cursor].ID
	}

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

	// Restore cursor by ID for stable navigation under async updates
	if m.cursorID != "" {
		for i, r := range m.visibleResults {
			if r.ID == m.cursorID {
				m.cursor = i
				m.clampOffset()
				return
			}
		}
	}

	// Fallback: clamp cursor
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
	title := styleHeader.Render("depsclean")
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

	// Details panel
	if m.showDetails && m.cursor >= 0 && m.cursor < len(m.visibleResults) {
		r := m.visibleResults[m.cursor]
		sb.WriteString(styleSeparator.Render(strings.Repeat("─", w)))
		sb.WriteString("\n")
		sb.WriteString(RenderDetailsPanel(r, w))
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
		m.dryRun,
		w,
	))

	return sb.String()
}
