package config

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
