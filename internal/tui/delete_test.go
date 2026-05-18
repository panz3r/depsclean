// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/panz3r/depsclean/internal/delete"
	"github.com/panz3r/depsclean/internal/model"
)

// dryRunDeleter returns a Deleter suitable for tests (dry-run, node_modules target).
func dryRunDeleter() *delete.Deleter {
	return &delete.Deleter{DryRun: true, Targets: []string{"node_modules"}}
}

// modelWithResults builds a model pre-loaded with the given results.
func modelWithResults(results ...model.Result) Model {
	m := newModel(nil)
	for _, r := range results {
		m.addResult(r)
	}
	return m
}

// --- toggleSelectAll ---

func TestToggleSelectAll_SelectsAllEligible(t *testing.T) {
	m := modelWithResults(
		makeResult("/a/node_modules", 100, model.StatusReady),
		makeResult("/b/node_modules", 200, model.StatusReady),
	)
	m.toggleSelectAll()
	for _, r := range m.visibleResults {
		if !m.selected[r.ID] {
			t.Errorf("expected %q to be selected", r.ID)
		}
	}
}

func TestToggleSelectAll_DeselectsWhenAllSelected(t *testing.T) {
	m := modelWithResults(
		makeResult("/a/node_modules", 100, model.StatusReady),
		makeResult("/b/node_modules", 200, model.StatusReady),
	)
	// First call: select all.
	m.toggleSelectAll()
	// Second call: deselect all.
	m.toggleSelectAll()
	if len(m.selected) != 0 {
		t.Errorf("expected 0 selected after second toggle, got %d", len(m.selected))
	}
}

func TestToggleSelectAll_SkipsDeletedRows(t *testing.T) {
	m := modelWithResults(
		makeResult("/a/node_modules", 100, model.StatusReady),
		makeResult("/b/node_modules", 200, model.StatusDeleted),
	)
	m.toggleSelectAll()
	if m.selected["/b/node_modules"] {
		t.Error("deleted row should not be selected by select-all")
	}
	if !m.selected["/a/node_modules"] {
		t.Error("ready row should be selected by select-all")
	}
}

// --- handleRangeSelect ---

func TestHandleRangeSelect_SetsAnchor(t *testing.T) {
	m := modelWithResults(
		makeResult("/a/node_modules", 300, model.StatusReady),
		makeResult("/b/node_modules", 200, model.StatusReady),
		makeResult("/c/node_modules", 100, model.StatusReady),
	)
	m.cursor = 0
	m.handleRangeSelect()
	if m.rangeAnchor == "" {
		t.Error("expected rangeAnchor to be set after first press")
	}
}

func TestHandleRangeSelect_SelectsRange(t *testing.T) {
	m := modelWithResults(
		makeResult("/a/node_modules", 300, model.StatusReady),
		makeResult("/b/node_modules", 200, model.StatusReady),
		makeResult("/c/node_modules", 100, model.StatusReady),
	)
	// Visible order is size-desc: a(300), b(200), c(100)
	m.cursor = 0
	m.handleRangeSelect() // sets anchor at index 0

	m.cursor = 2
	m.handleRangeSelect() // selects range [0..2]

	for _, r := range m.visibleResults {
		if !m.selected[r.ID] {
			t.Errorf("expected %q to be selected in range", r.ID)
		}
	}
}

func TestHandleRangeSelect_ResetsAnchor(t *testing.T) {
	m := modelWithResults(
		makeResult("/a/node_modules", 300, model.StatusReady),
		makeResult("/b/node_modules", 200, model.StatusReady),
	)
	m.cursor = 0
	m.handleRangeSelect() // set anchor
	m.cursor = 1
	m.handleRangeSelect() // apply range
	if m.rangeAnchor != "" {
		t.Errorf("expected rangeAnchor to be cleared, got %q", m.rangeAnchor)
	}
}

// --- Delete single (x key) ---

func TestDeleteSingle_TransitionsToDeleting(t *testing.T) {
	// Create a real temp directory named node_modules so Deleter can stat it.
	tmp := t.TempDir()
	dirPath := filepath.Join(tmp, "node_modules")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	r := model.Result{
		ID:          dirPath,
		Path:        dirPath,
		ProjectPath: tmp,
		Basename:    "node_modules",
		Status:      model.StatusReady,
	}
	m := newModel(nil)
	m.deleter = dryRunDeleter()
	m.addResult(r)

	// Send the "x" key.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := updated.(Model)

	if m2.allResults[0].Status != model.StatusDeleting {
		t.Errorf("expected StatusDeleting, got %v", m2.allResults[0].Status)
	}
}

// --- applyDeleteResult ---

func TestApplyDeleteResult_Success_TransitionsToDeleted(t *testing.T) {
	m := modelWithResults(makeResult("/a/node_modules", 100, model.StatusDeleting))

	m.applyDeleteResult(deleteResultMsg{id: "/a/node_modules", err: nil})

	if m.allResults[0].Status != model.StatusDeleted {
		t.Errorf("expected StatusDeleted, got %v", m.allResults[0].Status)
	}
	if m.allResults[0].ErrorMsg != "" {
		t.Errorf("expected empty ErrorMsg, got %q", m.allResults[0].ErrorMsg)
	}
}

func TestApplyDeleteResult_Error_TransitionsToError(t *testing.T) {
	m := modelWithResults(makeResult("/a/node_modules", 100, model.StatusDeleting))

	m.applyDeleteResult(deleteResultMsg{id: "/a/node_modules", err: &errMsg{"disk full"}})

	if m.allResults[0].Status != model.StatusError {
		t.Errorf("expected StatusError, got %v", m.allResults[0].Status)
	}
	if m.allResults[0].ErrorMsg != "disk full" {
		t.Errorf("expected ErrorMsg %q, got %q", "disk full", m.allResults[0].ErrorMsg)
	}
}

// --- deleteSelected ---

func TestDeleteSelected_BatchesAllSelected(t *testing.T) {
	tmp := t.TempDir()

	// Create two real node_modules directories.
	paths := []string{
		filepath.Join(tmp, "proj1", "node_modules"),
		filepath.Join(tmp, "proj2", "node_modules"),
	}
	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	m := newModel(nil)
	m.deleter = dryRunDeleter()
	for _, p := range paths {
		r := model.Result{
			ID:          p,
			Path:        p,
			ProjectPath: filepath.Dir(p),
			Basename:    "node_modules",
			Status:      model.StatusReady,
		}
		m.addResult(r)
		m.selected[p] = true
	}

	cmd := m.deleteSelected()
	if cmd == nil {
		t.Fatal("expected a batch Cmd, got nil")
	}

	// All selected entries should now be StatusDeleting.
	for _, r := range m.allResults {
		if r.Status != model.StatusDeleting {
			t.Errorf("expected %q to be StatusDeleting, got %v", r.ID, r.Status)
		}
	}

	// selected map should be cleared.
	if len(m.selected) != 0 {
		t.Errorf("expected selected to be cleared, got %d entries", len(m.selected))
	}
}

type errMsg struct{ msg string }

func (e *errMsg) Error() string { return e.msg }
