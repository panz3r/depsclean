// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Version is set at build time via -ldflags; defaults to "dev".
var Version = "dev"

// CheckResult holds the result of an update check.
type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	Err             error
}

// Check queries the GitHub Releases API for panz3r/depsclean and returns
// the latest release tag. The request is made with a short timeout (5 s).
// If Version == "dev", returns CheckResult{CurrentVersion: "dev"} without making an HTTP request.
func Check(ctx context.Context) CheckResult {
	result := CheckResult{CurrentVersion: Version}
	if Version == "dev" {
		return result
	}

	// Use 5-second deadline unless caller already set a shorter one.
	maxDeadline := time.Now().Add(5 * time.Second)
	if deadline, ok := ctx.Deadline(); !ok || deadline.After(maxDeadline) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, maxDeadline)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/panz3r/depsclean/releases/latest", nil)
	if err != nil {
		result.Err = err
		return result
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		result.Err = err
		return result
	}

	latest := strings.TrimPrefix(payload.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	result.LatestVersion = latest
	result.UpdateAvailable = latest != "" && latest != current
	return result
}

// FormatNotice returns a human-readable update notice string,
// or "" if no update is available or the check failed.
func FormatNotice(r CheckResult) string {
	if r.Err != nil || !r.UpdateAvailable {
		return ""
	}
	return fmt.Sprintf("A new version is available: v%s (current: v%s)\nDownload: https://github.com/panz3r/depsclean/releases/latest",
		r.LatestVersion, r.CurrentVersion)
}
