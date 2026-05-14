package scan

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/panz3r/depsclean/internal/model"
	"github.com/panz3r/depsclean/internal/platform"
)

// Analyze enriches DiscoveredEvents from in with size, age, and package metadata.
// Non-DiscoveredEvents (including DoneEvent) are forwarded after all analysis work completes.
// workers controls concurrency; 0 or negative uses a default of 8.
func Analyze(ctx context.Context, in <-chan Event, workers int) <-chan Event {
	if workers <= 0 {
		workers = 8
	}

	out := make(chan Event, 64)
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	go func() {
		var doneEv *DoneEvent

	loop:
		for {
			// Priority check: always honour context cancellation before blocking on input.
			select {
			case <-ctx.Done():
				out <- ErrorEvent{Err: ctx.Err()}
				break loop
			default:
			}

			select {
			case <-ctx.Done():
				out <- ErrorEvent{Err: ctx.Err()}
				break loop
			case ev, ok := <-in:
				if !ok {
					break loop
				}
				switch e := ev.(type) {
				case DiscoveredEvent:
					sem <- struct{}{}
					wg.Add(1)
					go func(r model.Result) {
						defer func() {
							<-sem
							wg.Done()
						}()
						out <- AnalyzedEvent{Result: analyzeResult(ctx, r)}
					}(e.Result)
				case DoneEvent:
					doneEv = &e
					break loop
				default:
					out <- ev
				}
			}
		}

		wg.Wait()
		if doneEv != nil {
			out <- *doneEv
		}
		close(out)
	}()

	return out
}

func analyzeResult(ctx context.Context, r model.Result) model.Result {
	size, err := platform.DiskUsage(r.Path)
	r.SizeBytes = size
	if err != nil {
		r.ErrorMsg = err.Error()
	}

	r.LastModified = findLastModified(r.Path)
	readPackageMetadata(&r)
	r.Status = model.StatusReady
	return r
}

func findLastModified(path string) time.Time {
	var latest time.Time
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err == nil && info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	return latest
}

type packageJSON struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	PackageManager string `json:"packageManager"`
}

func readPackageMetadata(r *model.Result) {
	pkgPath := filepath.Join(r.ProjectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}
	r.PackageName = pkg.Name
	r.PackageVersion = pkg.Version
	r.PackageManager = detectPackageManager(r.ProjectPath, pkg.PackageManager)
}

func detectPackageManager(projectPath, pkgManagerField string) model.PackageManager {
	// Check packageManager field first
	switch {
	case strings.HasPrefix(pkgManagerField, "yarn"):
		return model.PackageManagerYarn
	case strings.HasPrefix(pkgManagerField, "pnpm"):
		return model.PackageManagerPNPM
	case strings.HasPrefix(pkgManagerField, "npm"):
		return model.PackageManagerNPM
	case strings.HasPrefix(pkgManagerField, "bun"):
		return model.PackageManagerBun
	}

	// Check lockfiles
	lockfiles := []struct {
		name string
		pm   model.PackageManager
	}{
		{"bun.lockb", model.PackageManagerBun},
		{"bun.lock", model.PackageManagerBun},
		{"pnpm-lock.yaml", model.PackageManagerPNPM},
		{"yarn.lock", model.PackageManagerYarn},
		{"package-lock.json", model.PackageManagerNPM},
		{"npm-shrinkwrap.json", model.PackageManagerNPM},
	}

	for _, lf := range lockfiles {
		if _, err := os.Stat(filepath.Join(projectPath, lf.name)); err == nil {
			return lf.pm
		}
	}

	return model.PackageManagerUnknown
}
