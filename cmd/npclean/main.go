package main

import (
"context"
"fmt"
"os"
"strings"
"time"

"github.com/panz3r/npclean/internal/app"
"github.com/panz3r/npclean/internal/config"
"github.com/panz3r/npclean/internal/update"
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
root.AddCommand(buildVersionCmd())
root.AddCommand(buildProfilesCmd())
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

// Local flag variables for Changed detection
var flagRoot string
var flagProfile string
var flagExcludes []string
var flagSkipHidden bool
var flagMaxDepth int
var flagConfigFile string
var flagFormat string

cmd := &cobra.Command{
Use:   "scan",
Short: "Scan for dependency directories and print their paths",
RunE: func(cmd *cobra.Command, args []string) error {
// Step 1: start with defaults (already in cfg)

// Step 2: auto-detect or use --config flag
configPath := flagConfigFile
if configPath == "" {
configPath = config.FindConfigFile()
}

// Step 3: load and apply file config
if configPath != "" {
fc, err := config.LoadFile(configPath)
if err != nil {
return fmt.Errorf("loading config file %q: %w", configPath, err)
}
config.ApplyFile(&cfg, fc)
}

// Step 4: apply CLI flags only if explicitly set by the user
if cmd.Flags().Changed("root") {
cfg.RootDir = flagRoot
}
if cmd.Flags().Changed("profile") {
cfg.Profile = flagProfile
}
if cmd.Flags().Changed("exclude") {
cfg.Excludes = flagExcludes
}
if cmd.Flags().Changed("skip-hidden") {
cfg.SkipHidden = flagSkipHidden
}
if cmd.Flags().Changed("max-depth") {
cfg.MaxDepth = flagMaxDepth
}
if cmd.Flags().Changed("format") {
cfg.OutputFormat = config.OutputFormat(flagFormat)
}

// Apply profile targets
if p, ok := config.LookupProfile(cfg.Profile); ok {
cfg.Targets = p.Targets
}

return app.Run(cfg)
},
}

cmd.Flags().StringVar(&flagRoot, "root", cfg.RootDir, "Root directory to scan")
cmd.Flags().StringVar(&flagProfile, "profile", cfg.Profile, "Built-in profile to use (e.g. node)")
cmd.Flags().StringArrayVar(&flagExcludes, "exclude", cfg.Excludes, "Patterns to exclude (can be repeated)")
cmd.Flags().BoolVar(&flagSkipHidden, "skip-hidden", cfg.SkipHidden, "Skip hidden directories")
cmd.Flags().IntVar(&flagMaxDepth, "max-depth", cfg.MaxDepth, "Maximum directory depth to scan (0=unlimited)")
cmd.Flags().StringVar(&flagConfigFile, "config", "", "Explicit config file path (bypasses auto-detection)")
cmd.Flags().StringVar(&flagFormat, "format", string(cfg.OutputFormat), "Output format: text, json, ndjson")

return cmd
}

func buildVersionCmd() *cobra.Command {
checkUpdate := false
cmd := &cobra.Command{
Use:   "version",
Short: "Print version information",
RunE: func(cmd *cobra.Command, args []string) error {
fmt.Printf("npclean %s\n", update.Version)
if checkUpdate {
ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
defer cancel()
result := update.Check(ctx)
if notice := update.FormatNotice(result); notice != "" {
fmt.Println(notice)
}
}
return nil
},
}
cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "Check for a newer version on GitHub")
return cmd
}

func buildProfilesCmd() *cobra.Command {
return &cobra.Command{
Use:   "profiles",
Short: "List built-in scan profiles",
Run: func(cmd *cobra.Command, args []string) {
for _, name := range config.ListProfiles() {
p := config.BuiltinProfiles[name]
fmt.Printf("  %-10s  %s\n    targets: %s\n\n", p.Name, p.Description, strings.Join(p.Targets, ", "))
}
},
}
}
