package app

import (
	"context"
	"fmt"
	"os"

	"github.com/panz3r/npclean/internal/config"
	"github.com/panz3r/npclean/internal/output"
	"github.com/panz3r/npclean/internal/scan"
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
