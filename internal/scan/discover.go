package scan

import (
	"os"
	"path/filepath"

	"github.com/panz3r/npclean/internal/config"
	"github.com/panz3r/npclean/internal/model"
)

// Discover walks rootDir and emits Events on the returned channel.
// The channel is closed when the walk completes.
func Discover(rootDir string, cfg config.Config) <-chan Event {
	ch := make(chan Event)
	targets := make(map[string]struct{}, len(cfg.Targets))
	for _, t := range cfg.Targets {
		targets[t] = struct{}{}
	}

	go func() {
		defer close(ch)
		total := 0

		err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				ch <- ErrorEvent{Err: err}
				return nil // keep walking
			}
			if !d.IsDir() {
				return nil
			}
			base := d.Name()
			// Skip hidden directories if configured
			if cfg.SkipHidden && len(base) > 0 && base[0] == '.' && path != rootDir {
				return filepath.SkipDir
			}
			if _, ok := targets[base]; ok {
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
