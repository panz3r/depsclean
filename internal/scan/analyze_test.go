package scan

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/panz3r/npclean/internal/model"
)

// nonSensitiveTempDirForAnalyze creates a temp dir in the working directory to
// avoid sensitive-path issues on macOS where os.TempDir() lands under /private/var/…
func nonSensitiveTempDirForAnalyze(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(wd, "testAnalyze-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// makeNodeModulesDir creates a temp dir structure simulating a project with node_modules.
// Returns (projectDir, nodeModulesDir).
func makeNodeModulesDir(t *testing.T, pkgJSON map[string]string, lockfile string) (string, string) {
	t.Helper()
	root := nonSensitiveTempDirForAnalyze(t)
	nmDir := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(nmDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write a dummy file so DiskUsage returns > 0
	if err := os.WriteFile(filepath.Join(nmDir, "pkg.js"), []byte("// dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	if pkgJSON != nil {
		data, _ := json.Marshal(pkgJSON)
		if err := os.WriteFile(filepath.Join(root, "package.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	if lockfile != "" {
		if err := os.WriteFile(filepath.Join(root, lockfile), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return root, nmDir
}

func feedAndCollect(ctx context.Context, results []model.Result) []Event {
	in := make(chan Event, len(results)+1)
	for _, r := range results {
		in <- DiscoveredEvent{Result: r}
	}
	in <- DoneEvent{Total: len(results)}
	close(in)

	out := Analyze(ctx, in, 2)
	var events []Event
	for ev := range out {
		events = append(events, ev)
	}
	return events
}

func TestAnalyze_SizePopulated(t *testing.T) {
	projectDir, nmDir := makeNodeModulesDir(t, nil, "")

	r := model.Result{
		ID:          nmDir,
		Path:        nmDir,
		ProjectPath: projectDir,
		Basename:    "node_modules",
		Status:      model.StatusPending,
	}

	events := feedAndCollect(context.Background(), []model.Result{r})

	var analyzed *AnalyzedEvent
	for _, ev := range events {
		if ae, ok := ev.(AnalyzedEvent); ok {
			analyzed = &ae
		}
	}
	if analyzed == nil {
		t.Fatal("expected AnalyzedEvent, got none")
	}
	if analyzed.Result.SizeBytes < 0 {
		t.Errorf("SizeBytes should be >= 0, got %d", analyzed.Result.SizeBytes)
	}
	if analyzed.Result.Status != model.StatusReady {
		t.Errorf("Status should be StatusReady, got %v", analyzed.Result.Status)
	}
}

func TestAnalyze_PackageMetadata(t *testing.T) {
	projectDir, nmDir := makeNodeModulesDir(t, map[string]string{
		"name":    "my-app",
		"version": "1.2.3",
	}, "")

	r := model.Result{
		ID:          nmDir,
		Path:        nmDir,
		ProjectPath: projectDir,
		Basename:    "node_modules",
		Status:      model.StatusPending,
	}

	events := feedAndCollect(context.Background(), []model.Result{r})

	var analyzed *AnalyzedEvent
	for _, ev := range events {
		if ae, ok := ev.(AnalyzedEvent); ok {
			analyzed = &ae
		}
	}
	if analyzed == nil {
		t.Fatal("expected AnalyzedEvent, got none")
	}
	if analyzed.Result.PackageName != "my-app" {
		t.Errorf("PackageName = %q, want %q", analyzed.Result.PackageName, "my-app")
	}
	if analyzed.Result.PackageVersion != "1.2.3" {
		t.Errorf("PackageVersion = %q, want %q", analyzed.Result.PackageVersion, "1.2.3")
	}
}

func TestAnalyze_PackageManagerFromLockfile(t *testing.T) {
	tests := []struct {
		lockfile string
		want     model.PackageManager
	}{
		{"package-lock.json", model.PackageManagerNPM},
		{"yarn.lock", model.PackageManagerYarn},
		{"pnpm-lock.yaml", model.PackageManagerPNPM},
		{"bun.lockb", model.PackageManagerBun},
	}

	for _, tt := range tests {
		t.Run(tt.lockfile, func(t *testing.T) {
			projectDir, nmDir := makeNodeModulesDir(t, map[string]string{"name": "test"}, tt.lockfile)
			r := model.Result{
				ID:          nmDir,
				Path:        nmDir,
				ProjectPath: projectDir,
				Basename:    "node_modules",
				Status:      model.StatusPending,
			}

			events := feedAndCollect(context.Background(), []model.Result{r})

			var analyzed *AnalyzedEvent
			for _, ev := range events {
				if ae, ok := ev.(AnalyzedEvent); ok {
					analyzed = &ae
				}
			}
			if analyzed == nil {
				t.Fatal("expected AnalyzedEvent, got none")
			}
			if analyzed.Result.PackageManager != tt.want {
				t.Errorf("PackageManager = %q, want %q", analyzed.Result.PackageManager, tt.want)
			}
		})
	}
}

func TestAnalyze_DoneEventIsLast(t *testing.T) {
	projectDir, nmDir := makeNodeModulesDir(t, nil, "")
	r := model.Result{
		ID:          nmDir,
		Path:        nmDir,
		ProjectPath: projectDir,
		Basename:    "node_modules",
		Status:      model.StatusPending,
	}

	events := feedAndCollect(context.Background(), []model.Result{r})

	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	last := events[len(events)-1]
	if _, ok := last.(DoneEvent); !ok {
		t.Errorf("last event should be DoneEvent, got %T", last)
	}
}

func TestAnalyze_ContextCancellation(t *testing.T) {
	// Pre-cancel the context so Analyze sees cancellation immediately.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Closed empty channel so any drain attempt also terminates promptly.
	in := make(chan Event)
	close(in)

	out := Analyze(ctx, in, 2)

	done := make(chan struct{})
	var events []Event
	go func() {
		for ev := range out {
			events = append(events, ev)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Analyze to complete after context cancellation")
	}

	// Verify an ErrorEvent carrying context.Canceled was emitted.
	var hasError bool
	for _, ev := range events {
		if ee, ok := ev.(ErrorEvent); ok && ee.Err == context.Canceled {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected ErrorEvent with context.Canceled after cancellation")
	}
}

func TestDetectPackageManager_Field(t *testing.T) {
	tests := []struct {
		field string
		want  model.PackageManager
	}{
		{"yarn@3.2.0", model.PackageManagerYarn},
		{"pnpm@8.0.0", model.PackageManagerPNPM},
		{"npm@9.0.0", model.PackageManagerNPM},
		{"bun@1.0.0", model.PackageManagerBun},
		{"", model.PackageManagerUnknown},
	}

	dir := nonSensitiveTempDirForAnalyze(t) // no lockfiles

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := detectPackageManager(dir, tt.field)
			if got != tt.want {
				t.Errorf("detectPackageManager(%q, %q) = %q, want %q", dir, tt.field, got, tt.want)
			}
		})
	}
}

func TestDetectPackageManager_Lockfiles(t *testing.T) {
	tests := []struct {
		lockfile string
		want     model.PackageManager
	}{
		{"bun.lockb", model.PackageManagerBun},
		{"bun.lock", model.PackageManagerBun},
		{"pnpm-lock.yaml", model.PackageManagerPNPM},
		{"yarn.lock", model.PackageManagerYarn},
		{"package-lock.json", model.PackageManagerNPM},
		{"npm-shrinkwrap.json", model.PackageManagerNPM},
	}

	for _, tt := range tests {
		t.Run(tt.lockfile, func(t *testing.T) {
			dir := nonSensitiveTempDirForAnalyze(t)
			if err := os.WriteFile(filepath.Join(dir, tt.lockfile), []byte(""), 0644); err != nil {
				t.Fatal(err)
			}
			got := detectPackageManager(dir, "")
			if got != tt.want {
				t.Errorf("detectPackageManager with %q = %q, want %q", tt.lockfile, got, tt.want)
			}
		})
	}
}
