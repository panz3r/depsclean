package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.RootDir != "." {
		t.Errorf("expected RootDir='.', got %q", cfg.RootDir)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0] != "node_modules" {
		t.Errorf("unexpected Targets: %v", cfg.Targets)
	}
	if !cfg.SkipHidden {
		t.Error("expected SkipHidden=true")
	}
	if cfg.MaxDepth != 10 {
		t.Errorf("expected MaxDepth=10, got %d", cfg.MaxDepth)
	}
	if cfg.DryRun {
		t.Error("expected DryRun=false")
	}
	if cfg.OutputFormat != OutputFormatText {
		t.Errorf("expected OutputFormat=text, got %q", cfg.OutputFormat)
	}
	if cfg.Profile != "node" {
		t.Errorf("expected Profile=node, got %q", cfg.Profile)
	}
}

func TestApplyFile_OverridesDefaults(t *testing.T) {
	cfg := Default()
	profile := "python"
	maxDepth := 5
	dryRun := true
	fc := FileConfig{
		Profile:  &profile,
		MaxDepth: &maxDepth,
		DryRun:   &dryRun,
	}
	ApplyFile(&cfg, fc)
	if cfg.Profile != "python" {
		t.Errorf("expected Profile=python, got %q", cfg.Profile)
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("expected MaxDepth=5, got %d", cfg.MaxDepth)
	}
	if !cfg.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestApplyFile_NilFieldsPreserved(t *testing.T) {
	cfg := Default()
	original := cfg
	fc := FileConfig{} // all nil
	ApplyFile(&cfg, fc)
	if cfg.Profile != original.Profile {
		t.Errorf("Profile should be unchanged: got %q", cfg.Profile)
	}
	if cfg.MaxDepth != original.MaxDepth {
		t.Errorf("MaxDepth should be unchanged: got %d", cfg.MaxDepth)
	}
	if cfg.DryRun != original.DryRun {
		t.Errorf("DryRun should be unchanged: got %v", cfg.DryRun)
	}
}

func TestFindConfigFile_NotFound(t *testing.T) {
	// Use a fresh temp dir as cwd
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	// Temporarily override home to avoid picking up real user config
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	got := FindConfigFile()
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindConfigFile_FoundInCwd(t *testing.T) {
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	// Override HOME so the home-dir candidate is not accidentally found.
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	// Place .depsclean.json in the temp dir (now cwd).
	cfgPath := filepath.Join(tmp, ".depsclean.json")
	if err := os.WriteFile(cfgPath, []byte(`{"version":1}`), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got := FindConfigFile()
	if got == "" {
		t.Fatal("expected a non-empty path, got empty string")
	}
	if _, err := os.Stat(got); err != nil {
		t.Errorf("returned path does not exist: %v", err)
	}
}

func TestLoadFile_ValidFixture(t *testing.T) {
	path := filepath.Join("testdata", "valid.depsclean.json")
	fc, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if fc.Version != 1 {
		t.Errorf("Version = %d, want 1", fc.Version)
	}
	if fc.Profile == nil || *fc.Profile != "python" {
		t.Errorf("Profile = %v, want python", fc.Profile)
	}
	if len(fc.Targets) != 2 || fc.Targets[0] != ".venv" {
		t.Errorf("Targets = %v, want [.venv __pycache__]", fc.Targets)
	}
	if len(fc.Excludes) != 2 || fc.Excludes[0] != "vendor" {
		t.Errorf("Excludes = %v, want [vendor .git]", fc.Excludes)
	}
	if fc.SkipHidden == nil || *fc.SkipHidden != false {
		t.Errorf("SkipHidden = %v, want false", fc.SkipHidden)
	}
	if fc.MaxDepth == nil || *fc.MaxDepth != 5 {
		t.Errorf("MaxDepth = %v, want 5", fc.MaxDepth)
	}
	if fc.DryRun == nil || *fc.DryRun != true {
		t.Errorf("DryRun = %v, want true", fc.DryRun)
	}
	if fc.OutputFormat == nil || *fc.OutputFormat != "json" {
		t.Errorf("OutputFormat = %v, want json", fc.OutputFormat)
	}
}

func TestLoadFile_InvalidJSON(t *testing.T) {
	path := filepath.Join("testdata", "invalid.depsclean.json")
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadFile_NotFound(t *testing.T) {
	_, err := LoadFile(filepath.Join("testdata", "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestApplyFile_AllFields(t *testing.T) {
	cfg := Default()

	profile := "rust"
	targets := []string{"target"}
	excludes := []string{"build", "dist"}
	skipHidden := false
	maxDepth := 3
	dryRun := true
	outputFormat := "ndjson"

	fc := FileConfig{
		Profile:      &profile,
		Targets:      targets,
		Excludes:     excludes,
		SkipHidden:   &skipHidden,
		MaxDepth:     &maxDepth,
		DryRun:       &dryRun,
		OutputFormat: &outputFormat,
	}
	ApplyFile(&cfg, fc)

	if cfg.Profile != "rust" {
		t.Errorf("Profile = %q, want rust", cfg.Profile)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0] != "target" {
		t.Errorf("Targets = %v, want [target]", cfg.Targets)
	}
	if len(cfg.Excludes) != 2 || cfg.Excludes[0] != "build" {
		t.Errorf("Excludes = %v, want [build dist]", cfg.Excludes)
	}
	if cfg.SkipHidden != false {
		t.Errorf("SkipHidden = %v, want false", cfg.SkipHidden)
	}
	if cfg.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", cfg.MaxDepth)
	}
	if !cfg.DryRun {
		t.Error("DryRun = false, want true")
	}
	if cfg.OutputFormat != OutputFormatNDJSON {
		t.Errorf("OutputFormat = %q, want ndjson", cfg.OutputFormat)
	}
}

func TestApplyFile_LoadAndApply_RoundTrip(t *testing.T) {
	cfg := Default()
	path := filepath.Join("testdata", "valid.depsclean.json")
	fc, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	ApplyFile(&cfg, fc)

	// All fields from the fixture should be reflected in cfg.
	if cfg.Profile != "python" {
		t.Errorf("Profile = %q, want python", cfg.Profile)
	}
	if len(cfg.Targets) != 2 {
		t.Errorf("Targets len = %d, want 2", len(cfg.Targets))
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("MaxDepth = %d, want 5", cfg.MaxDepth)
	}
	if !cfg.DryRun {
		t.Error("DryRun should be true after applying fixture")
	}
	if cfg.OutputFormat != OutputFormatJSON {
		t.Errorf("OutputFormat = %q, want json", cfg.OutputFormat)
	}
}
