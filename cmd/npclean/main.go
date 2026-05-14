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
	root := &cobra.Command{
		Use:   "npclean",
		Short: "Clean up node_modules directories to reclaim disk space",
	}

	root.AddCommand(buildScanCmd())
	return root
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

	return cmd
}
