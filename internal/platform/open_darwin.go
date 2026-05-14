//go:build darwin

package platform

import (
	"fmt"
	"os/exec"
)

// Open opens path in the macOS Finder.
func Open(path string) error {
	cmd := exec.Command("open", path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
