package app

import (
	"fmt"

	"github.com/panz3r/npclean/internal/config"
	"github.com/panz3r/npclean/internal/scan"
)

// Run executes the scan pipeline with the given configuration and prints
// discovered paths to stdout.
func Run(cfg config.Config) error {
	events := scan.Discover(cfg.RootDir, cfg)
	for ev := range events {
		switch e := ev.(type) {
		case scan.DiscoveredEvent:
			fmt.Println(e.Result.Path)
		case scan.ErrorEvent:
			fmt.Printf("warn: %v\n", e.Err)
		case scan.DoneEvent:
			fmt.Printf("done: found %d target(s)\n", e.Total)
		}
	}
	return nil
}
