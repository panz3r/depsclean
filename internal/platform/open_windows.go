//go:build windows

// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
