// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package scan

import "github.com/panz3r/depsclean/internal/model"

// Event is the base type for all scan pipeline events.
type Event interface{ scanEvent() }

// DiscoveredEvent is emitted when a target directory is found.
type DiscoveredEvent struct{ Result model.Result }

// AnalyzedEvent is emitted when size/metadata analysis completes for a result.
type AnalyzedEvent struct{ Result model.Result }

// ErrorEvent is emitted when a non-fatal scan error occurs.
type ErrorEvent struct{ Err error }

// DoneEvent signals that scanning has completed.
type DoneEvent struct{ Total int }

func (DiscoveredEvent) scanEvent() {}
func (AnalyzedEvent) scanEvent()   {}
func (ErrorEvent) scanEvent()      {}
func (DoneEvent) scanEvent()       {}
