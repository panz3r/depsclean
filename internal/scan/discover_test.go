package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/panz3r/npclean/internal/config"
)

// nonSensitiveTempDir creates a temp dir in the current working directory so
// it is never filtered by IsSensitivePath. On macOS, os.TempDir() / t.TempDir()
// resolve to /private/var/folders/… which is a known sensitive path.
func nonSensitiveTempDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(wd, "testDiscover-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// collectEvents drains the channel and returns all events as a slice.
func collectEvents(ch <-chan Event) []Event {
	var events []Event
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}

func discoveredPaths(events []Event) []string {
	var paths []string
	for _, ev := range events {
		if de, ok := ev.(DiscoveredEvent); ok {
			paths = append(paths, de.Result.Path)
		}
	}
	return paths
}

func TestDiscover_Basic(t *testing.T) {
	root := nonSensitiveTempDir(t)
	// Create: root/project/node_modules
	projectDir := filepath.Join(root, "project")
	nmDir := filepath.Join(projectDir, "node_modules")
	if err := os.MkdirAll(nmDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Targets:    []string{"node_modules"},
		SkipHidden: false,
		MaxDepth:   0,
	}
	events := collectEvents(Discover(root, cfg))

	paths := discoveredPaths(events)
	if len(paths) != 1 {
		t.Fatalf("expected 1 discovered path, got %d: %v", len(paths), paths)
	}
	if paths[0] != nmDir {
		t.Errorf("expected %q, got %q", nmDir, paths[0])
	}

	// Last event should be DoneEvent with total=1
	last := events[len(events)-1]
	done, ok := last.(DoneEvent)
	if !ok {
		t.Fatalf("last event should be DoneEvent, got %T", last)
	}
	if done.Total != 1 {
		t.Errorf("DoneEvent.Total = %d, want 1", done.Total)
	}
}

func TestDiscover_SkipHidden(t *testing.T) {
	root := nonSensitiveTempDir(t)
	// Create: root/.hidden/node_modules (should be skipped)
	hiddenDir := filepath.Join(root, ".hidden")
	nmInHidden := filepath.Join(hiddenDir, "node_modules")
	if err := os.MkdirAll(nmInHidden, 0755); err != nil {
		t.Fatal(err)
	}
	// Create: root/visible/node_modules (should be found)
	visibleDir := filepath.Join(root, "visible")
	nmVisible := filepath.Join(visibleDir, "node_modules")
	if err := os.MkdirAll(nmVisible, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Targets:    []string{"node_modules"},
		SkipHidden: true,
		MaxDepth:   0,
	}
	events := collectEvents(Discover(root, cfg))
	paths := discoveredPaths(events)

	if len(paths) != 1 {
		t.Fatalf("expected 1 discovered path, got %d: %v", len(paths), paths)
	}
	if paths[0] != nmVisible {
		t.Errorf("expected %q, got %q", nmVisible, paths[0])
	}
}

func TestDiscover_ExcludePatterns(t *testing.T) {
	root := nonSensitiveTempDir(t)
	// Create: root/excluded/node_modules (should be excluded)
	excludedDir := filepath.Join(root, "excluded")
	nmExcluded := filepath.Join(excludedDir, "node_modules")
	if err := os.MkdirAll(nmExcluded, 0755); err != nil {
		t.Fatal(err)
	}
	// Create: root/included/node_modules (should be found)
	includedDir := filepath.Join(root, "included")
	nmIncluded := filepath.Join(includedDir, "node_modules")
	if err := os.MkdirAll(nmIncluded, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Targets:    []string{"node_modules"},
		SkipHidden: false,
		Excludes:   []string{"excluded"},
		MaxDepth:   0,
	}
	events := collectEvents(Discover(root, cfg))
	paths := discoveredPaths(events)

	if len(paths) != 1 {
		t.Fatalf("expected 1 discovered path, got %d: %v", len(paths), paths)
	}
	if paths[0] != nmIncluded {
		t.Errorf("expected %q, got %q", nmIncluded, paths[0])
	}
}

func TestDiscover_DepthLimit(t *testing.T) {
	root := nonSensitiveTempDir(t)
	// Create: root/a/node_modules (depth=2) - should be found at MaxDepth=3
	aDir := filepath.Join(root, "a")
	nmShallow := filepath.Join(aDir, "node_modules")
	if err := os.MkdirAll(nmShallow, 0755); err != nil {
		t.Fatal(err)
	}
	// Create: root/a/b/c/node_modules - parent c is at depth=3 which >= MaxDepth=3,
	// so c is skipped and node_modules is never visited.
	deepDir := filepath.Join(root, "a", "b", "c")
	nmDeep := filepath.Join(deepDir, "node_modules")
	if err := os.MkdirAll(nmDeep, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Targets:    []string{"node_modules"},
		SkipHidden: false,
		MaxDepth:   3,
	}
	events := collectEvents(Discover(root, cfg))
	paths := discoveredPaths(events)

	if len(paths) != 1 {
		t.Fatalf("expected 1 discovered path (shallow), got %d: %v", len(paths), paths)
	}
	if paths[0] != nmShallow {
		t.Errorf("expected %q, got %q", nmShallow, paths[0])
	}
}

func TestDiscover_MultipleTargets(t *testing.T) {
	root := nonSensitiveTempDir(t)

	targets := []string{"node_modules", "vendor"}
	dirs := []string{
		filepath.Join(root, "proj1", "node_modules"),
		filepath.Join(root, "proj2", "vendor"),
		filepath.Join(root, "proj3", "node_modules"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := config.Config{
		Targets:    targets,
		SkipHidden: false,
		MaxDepth:   0,
	}
	events := collectEvents(Discover(root, cfg))
	paths := discoveredPaths(events)

	if len(paths) != 3 {
		t.Fatalf("expected 3 discovered paths, got %d: %v", len(paths), paths)
	}
}
