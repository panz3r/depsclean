package config

// Profile defines a named set of scan targets with metadata.
type Profile struct {
	Name        string
	Description string
	Targets     []string
}

// BuiltinProfiles is the registry of built-in profiles.
var BuiltinProfiles = map[string]Profile{
	"node": {
		Name:        "node",
		Description: "Node.js – removes node_modules directories",
		Targets:     []string{"node_modules"},
	},
}

// LookupProfile returns the Profile for name, and a bool indicating if it was found.
func LookupProfile(name string) (Profile, bool) {
	p, ok := BuiltinProfiles[name]
	return p, ok
}
