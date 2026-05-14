package scan

import (
	"os"
	"path/filepath"

	"github.com/panz3r/depsclean/internal/config"
	"github.com/panz3r/depsclean/internal/model"
)

// Discover walks rootDir and emits Events on the returned channel.
// The channel is closed when the walk completes.
func Discover(rootDir string, cfg config.Config) <-chan Event {
	ch := make(chan Event, 64)

	go func() {
		defer close(ch)

		absRoot, err := filepath.Abs(rootDir)
		if err != nil {
			ch <- ErrorEvent{Err: err}
			ch <- DoneEvent{Total: 0}
			return
		}

		targets := make(map[string]struct{}, len(cfg.Targets))
		for _, t := range cfg.Targets {
			targets[t] = struct{}{}
		}

		total := 0

		err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				ch <- ErrorEvent{Err: err}
				return nil // keep walking
			}
			if !d.IsDir() {
				return nil
			}

			// Skip root itself
			if path == absRoot {
				return nil
			}

			base := d.Name()

			// Filter: hidden
			if cfg.SkipHidden && IsHidden(base) {
				return filepath.SkipDir
			}

			// Filter: excluded
			if IsExcluded(path, cfg.Excludes) {
				return filepath.SkipDir
			}

			// Filter: sensitive
			if IsSensitivePath(path) {
				return filepath.SkipDir
			}

			// Check if it's a target
			_, isTarget := targets[base]

			// Filter: depth (only skip non-targets at max depth)
			if cfg.MaxDepth > 0 {
				depth := PathDepth(path, absRoot)
				if depth >= cfg.MaxDepth && !isTarget {
					return filepath.SkipDir
				}
			}

			// Emit discovery
			if isTarget {
				r := model.Result{
					ID:          path,
					Path:        path,
					ProjectPath: filepath.Dir(path),
					Basename:    base,
					Status:      model.StatusPending,
				}
				ch <- DiscoveredEvent{Result: r}
				total++
				return filepath.SkipDir // don't descend into target
			}

			return nil
		})

		if err != nil {
			ch <- ErrorEvent{Err: err}
		}
		ch <- DoneEvent{Total: total}
	}()

	return ch
}
