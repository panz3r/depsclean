// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/panz3r/depsclean/internal/config"
	"github.com/panz3r/depsclean/internal/delete"
	"github.com/panz3r/depsclean/internal/model"
	"github.com/panz3r/depsclean/internal/output"
	"github.com/panz3r/depsclean/internal/scan"
)

// Run executes the scan pipeline with the given configuration and prints
// discovered paths to stdout according to cfg.OutputFormat.
func Run(cfg config.Config) error {
	ctx := context.Background()
	discoveries := scan.Discover(cfg.RootDir, cfg)
	events := scan.Analyze(ctx, discoveries, 0)

	switch cfg.OutputFormat {
	case config.OutputFormatNDJSON:
		w := output.NewNDJSONWriter(os.Stdout)
		for ev := range events {
			switch e := ev.(type) {
			case scan.AnalyzedEvent:
				if err := w.Write(e.Result); err != nil {
					return err
				}
			}
		}
	case config.OutputFormatJSON:
		w := output.NewJSONWriter(os.Stdout)
		for ev := range events {
			switch e := ev.(type) {
			case scan.AnalyzedEvent:
				w.Add(e.Result)
			}
		}
		return w.Flush()
	default: // text
		for ev := range events {
			switch e := ev.(type) {
			case scan.DiscoveredEvent:
				fmt.Printf("found: %s\n", e.Result.Path)
			case scan.AnalyzedEvent:
				fmt.Printf("sized: %s (%d bytes, pm=%s)\n", e.Result.Path, e.Result.SizeBytes, e.Result.PackageManager)
			case scan.ErrorEvent:
				fmt.Printf("warn: %v\n", e.Err)
			case scan.DoneEvent:
				fmt.Printf("done: found %d target(s)\n", e.Total)
			}
		}
	}
	return nil
}

// fmtBytes formats a byte count as a human-readable string.
func fmtBytes(b int64) string {
	const (
		GB = 1 << 30
		MB = 1 << 20
		KB = 1 << 10
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// RunDeleteAll scans for dependency directories and deletes them.
// confirmed must be true (or cfg.DryRun must be true) to perform actual deletions.
func RunDeleteAll(cfg config.Config, confirmed bool) error {
	ctx := context.Background()
	discoveries := scan.Discover(cfg.RootDir, cfg)
	events := scan.Analyze(ctx, discoveries, 0)

	var results []model.Result
	for ev := range events {
		if e, ok := ev.(scan.AnalyzedEvent); ok {
			results = append(results, e.Result)
		}
	}

	if len(results) == 0 {
		fmt.Println("No dependency directories found.")
		return nil
	}

	var totalBytes int64
	for _, r := range results {
		totalBytes += r.SizeBytes
	}
	fmt.Printf("Found %d directories totalling %s\n", len(results), fmtBytes(totalBytes))

	if cfg.DryRun {
		fmt.Println("(dry-run: no files will be deleted)")
	}

	if !confirmed && !cfg.DryRun {
		fmt.Println("Use --yes to confirm deletion.")
		return nil
	}

	targets := cfg.Targets
	if len(targets) == 0 {
		targets = []string{"node_modules"}
	}
	d := &delete.Deleter{DryRun: cfg.DryRun, Targets: targets}

	var deleted, failed int
	var firstErr error
	for _, r := range results {
		err := d.Delete(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", r.Path, err)
			failed++
			if firstErr == nil {
				firstErr = err
			}
		} else {
			if cfg.DryRun {
				fmt.Printf("  ~ %s (%s)\n", r.Path, fmtBytes(r.SizeBytes))
			} else {
				fmt.Printf("  ✓ %s (%s)\n", r.Path, fmtBytes(r.SizeBytes))
			}
			deleted++
		}
	}

	fmt.Printf("\nSummary: %d deleted, %d failed\n", deleted, failed)
	if failed > 0 {
		return errors.New("some deletions failed")
	}
	return nil
}
