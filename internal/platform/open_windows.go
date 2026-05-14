//go:build windows

package platform

import (
	"fmt"
	"os/exec"
)

// Open opens path in Windows Explorer.
func Open(path string) error {
	cmd := exec.Command("explorer", path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
