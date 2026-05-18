// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// OutputFormat describes the output mode.
type OutputFormat string

const (
	OutputFormatText   OutputFormat = "text"
	OutputFormatJSON   OutputFormat = "json"
	OutputFormatNDJSON OutputFormat = "ndjson"
)

// Config holds the resolved runtime configuration.
type Config struct {
	RootDir      string
	Targets      []string
	Excludes     []string
	SkipHidden   bool
	MaxDepth     int
	DryRun       bool
	OutputFormat OutputFormat
	Profile      string
}

// Default returns a Config populated with safe defaults.
func Default() Config {
	return Config{
		RootDir:      ".",
		Targets:      []string{"node_modules"},
		Excludes:     []string{},
		SkipHidden:   true,
		MaxDepth:     10,
		DryRun:       false,
		OutputFormat: OutputFormatText,
		Profile:      "node",
	}
}

// FileConfig holds optional configuration loaded from a JSON config file.
// Pointer fields allow distinguishing "not set" from zero-value.
type FileConfig struct {
	Version      int      `json:"version,omitempty"`
	Profile      *string  `json:"profile,omitempty"`
	Targets      []string `json:"targets,omitempty"`
	Excludes     []string `json:"excludes,omitempty"`
	SkipHidden   *bool    `json:"skip_hidden,omitempty"`
	MaxDepth     *int     `json:"max_depth,omitempty"`
	DryRun       *bool    `json:"dry_run,omitempty"`
	OutputFormat *string  `json:"output_format,omitempty"`
}

// FindConfigFile looks for .depsclean.json in cwd first, then ~/.config/depsclean/config.json.
// Returns the first found path or empty string.
func FindConfigFile() string {
	cwd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(cwd, ".depsclean.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidate := filepath.Join(home, ".config", "depsclean", "config.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// LoadFile reads and JSON-decodes the config file at path.
func LoadFile(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, err
	}
	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return FileConfig{}, err
	}
	return fc, nil
}

// ApplyFile applies non-nil file config fields onto cfg.
// File values override existing cfg values; nil pointer fields are not touched.
func ApplyFile(cfg *Config, fc FileConfig) {
	if fc.Profile != nil {
		cfg.Profile = *fc.Profile
	}
	if fc.Targets != nil {
		cfg.Targets = fc.Targets
	}
	if fc.Excludes != nil {
		cfg.Excludes = fc.Excludes
	}
	if fc.SkipHidden != nil {
		cfg.SkipHidden = *fc.SkipHidden
	}
	if fc.MaxDepth != nil {
		cfg.MaxDepth = *fc.MaxDepth
	}
	if fc.DryRun != nil {
		cfg.DryRun = *fc.DryRun
	}
	if fc.OutputFormat != nil {
		cfg.OutputFormat = OutputFormat(*fc.OutputFormat)
	}
}
