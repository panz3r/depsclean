package config

import (
	"os"
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
