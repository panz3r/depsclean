package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/panz3r/npclean/internal/scan"
)

// scanEventMsg carries one scan pipeline event and the channel for the next read.
type scanEventMsg struct {
	ev scan.Event
	ch <-chan scan.Event
}

// WaitForScanEvent returns a Cmd that blocks until the next event on ch, then
// returns a scanEventMsg. The caller's Update must schedule another WaitForScanEvent
// to keep consuming the channel.
func WaitForScanEvent(ch <-chan scan.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return scanEventMsg{ev: ev, ch: ch}
	}
}

// FixtureStartCmd returns a tea.Cmd that feeds all fixture results into the TUI
// as if they were live scan events (for demo/fallback mode).
func FixtureStartCmd() tea.Cmd {
	fixtures := FixtureResults()
	cmds := make([]tea.Cmd, 0, len(fixtures)+1)
	for _, r := range fixtures {
		r := r
		cmds = append(cmds, func() tea.Msg { return ResultMsg{R: r} })
	}
	cmds = append(cmds, func() tea.Msg { return ScanDoneMsg{Total: len(fixtures)} })
	return tea.Sequence(cmds...)
}
