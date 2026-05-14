//go:build !darwin && !windows

package platform

import (
	"fmt"
	"os/exec"
)

// Open opens path using xdg-open (Linux and other Unix-like systems).
func Open(path string) error {
	cmd := exec.Command("xdg-open", path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
