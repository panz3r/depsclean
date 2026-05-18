// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/panz3r/depsclean/internal/config"
	"github.com/panz3r/depsclean/internal/delete"
	"github.com/panz3r/depsclean/internal/scan"
	"github.com/panz3r/depsclean/internal/tui"
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
