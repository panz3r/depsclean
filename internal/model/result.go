package model

import "time"

// Status represents the lifecycle state of a discovered result.
type Status int

const (
	StatusPending  Status = iota // discovered, not yet sized
	StatusReady                  // fully analyzed, available for action
	StatusDeleting               // deletion in progress
	StatusDeleted                // successfully deleted
	StatusError                  // error during analysis or deletion
)

// PackageManager identifies the package manager used in the project.
type PackageManager string

const (
	PackageManagerUnknown PackageManager = ""
	PackageManagerNPM     PackageManager = "npm"
	PackageManagerYarn    PackageManager = "yarn"
	PackageManagerPNPM    PackageManager = "pnpm"
	PackageManagerBun     PackageManager = "bun"

	// Non-Node ecosystems
	PackageManagerPython PackageManager = "python"
	PackageManagerRust   PackageManager = "rust"
	PackageManagerGo     PackageManager = "go"
	PackageManagerPHP    PackageManager = "php"
	PackageManagerRuby   PackageManager = "ruby"
	PackageManagerJava   PackageManager = "java"
)

// Result represents a discovered dependency directory.
type Result struct {
	ID             string // absolute path (stable identity)
	Path           string // absolute path to the target dir
	ProjectPath    string // parent project directory
	Basename       string // target dir name (e.g. "node_modules")
	SizeBytes      int64
	LastModified   time.Time
	Status         Status
	PackageManager PackageManager
	PackageName    string
	PackageVersion string
	ErrorMsg       string
}
