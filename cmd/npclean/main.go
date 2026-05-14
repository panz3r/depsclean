package main

import (
	"fmt"
	"os"

	"github.com/panz3r/npclean/internal/app"
	"github.com/panz3r/npclean/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := buildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	cfg := config.Default()

	root := &cobra.Command{
		Use:   "npclean",
		Short: "Clean up node_modules directories to reclaim disk space",
		RunE: func(cmd *cobra.Command, args []string) error {
			if p, ok := config.LookupProfile(cfg.Profile); ok {
				cfg.Targets = p.Targets
			}
			return app.RunTUI(cfg)
		},
	}

	root.Flags().StringVar(&cfg.RootDir, "root", cfg.RootDir, "Root directory to scan")
	root.Flags().StringVar(&cfg.Profile, "profile", cfg.Profile, "Built-in profile to use (e.g. node)")
	root.Flags().StringArrayVar(&cfg.Excludes, "exclude", cfg.Excludes, "Patterns to exclude (can be repeated)")
	root.Flags().BoolVar(&cfg.SkipHidden, "skip-hidden", cfg.SkipHidden, "Skip hidden directories")
	root.Flags().IntVar(&cfg.MaxDepth, "max-depth", cfg.MaxDepth, "Maximum directory depth (0=unlimited)")

	root.AddCommand(buildScanCmd())
	root.AddCommand(buildUICmd())
	return root
}

func buildUICmd() *cobra.Command {
	cfg := config.Default()

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if p, ok := config.LookupProfile(cfg.Profile); ok {
				cfg.Targets = p.Targets
			}
			return app.RunTUI(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.RootDir, "root", cfg.RootDir, "Root directory to scan")
	cmd.Flags().StringVar(&cfg.Profile, "profile", cfg.Profile, "Built-in profile to use (e.g. node)")
	cmd.Flags().StringArrayVar(&cfg.Excludes, "exclude", cfg.Excludes, "Patterns to exclude (can be repeated)")
	cmd.Flags().BoolVar(&cfg.SkipHidden, "skip-hidden", cfg.SkipHidden, "Skip hidden directories")
	cmd.Flags().IntVar(&cfg.MaxDepth, "max-depth", cfg.MaxDepth, "Maximum directory depth (0=unlimited)")

	return cmd
}

func buildScanCmd() *cobra.Command {
	cfg := config.Default()

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for dependency directories and print their paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If a named profile is set, merge its targets into cfg.
			if p, ok := config.LookupProfile(cfg.Profile); ok {
				cfg.Targets = p.Targets
			}
			return app.Run(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.RootDir, "root", cfg.RootDir, "Root directory to scan")
	cmd.Flags().StringVar(&cfg.Profile, "profile", cfg.Profile, "Built-in profile to use (e.g. node)")
	cmd.Flags().StringArrayVar(&cfg.Excludes, "exclude", cfg.Excludes, "Patterns to exclude (can be repeated)")
	cmd.Flags().BoolVar(&cfg.SkipHidden, "skip-hidden", cfg.SkipHidden, "Skip hidden directories")
	cmd.Flags().IntVar(&cfg.MaxDepth, "max-depth", cfg.MaxDepth, "Maximum directory depth to scan (0=unlimited)")

	return cmd
}
