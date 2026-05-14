package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/panz3r/npclean/internal/config"
	"github.com/panz3r/npclean/internal/delete"
	"github.com/panz3r/npclean/internal/scan"
	"github.com/panz3r/npclean/internal/tui"
)

// RunTUI launches the interactive TUI connected to a live scan of cfg.
func RunTUI(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discoveries := scan.Discover(cfg.RootDir, cfg)
	events := scan.Analyze(ctx, discoveries, 0)

	startCmd := tui.WaitForScanEvent(events)
	targets := cfg.Targets
	if len(targets) == 0 {
		targets = []string{"node_modules"}
	}
	d := &delete.Deleter{DryRun: cfg.DryRun, Targets: targets}
	m := tui.NewWithScanAndDeleter(startCmd, d)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunTUIDemo launches the TUI with fixture data (no real scan).
func RunTUIDemo() error {
	m := tui.New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
