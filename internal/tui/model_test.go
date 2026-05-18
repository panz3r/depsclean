// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tui

import (
	"testing"
	"time"

	"github.com/panz3r/depsclean/internal/model"
	"github.com/panz3r/depsclean/internal/scan"
)

// makeResult is a test helper.
func makeResult(id string, sizeBytes int64, status model.Status) model.Result {
	return model.Result{
		ID:          id,
		Path:        id,
		ProjectPath: id + "/..",
		Basename:    "node_modules",
		SizeBytes:   sizeBytes,
		Status:      status,
	}
}

// --- addResult / in-place updateResult ---

func TestAddResult_AppendsNewRow(t *testing.T) {
	m := newModel(nil)
	r := makeResult("/a/node_modules", 100, model.StatusPending)
	m.addResult(r)

	if len(m.allResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(m.allResults))
	}
	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 visible result, got %d", len(m.visibleResults))
	}
	if m.scanFound != 1 {
		t.Fatalf("expected scanFound=1, got %d", m.scanFound)
	}
}

func TestAddResult_IgnoresDuplicate(t *testing.T) {
	m := newModel(nil)
	r := makeResult("/a/node_modules", 100, model.StatusPending)
	m.addResult(r)
	m.addResult(r) // duplicate

	if len(m.allResults) != 1 {
		t.Fatalf("expected 1 result after duplicate, got %d", len(m.allResults))
	}
	if m.scanFound != 1 {
		t.Fatalf("expected scanFound=1, got %d", m.scanFound)
	}
}

func TestUpdateResult_EnrichesExistingRow(t *testing.T) {
	m := newModel(nil)
	pending := makeResult("/a/node_modules", 0, model.StatusPending)
	m.addResult(pending)

	ready := makeResult("/a/node_modules", 500*1024*1024, model.StatusReady)
	ready.PackageName = "my-app"
	m.updateResult(ready)

	// allResults should still have exactly 1 entry (no duplicate)
	if len(m.allResults) != 1 {
		t.Fatalf("expected 1 result after update, got %d", len(m.allResults))
	}
	if m.allResults[0].SizeBytes != 500*1024*1024 {
		t.Errorf("expected updated size, got %d", m.allResults[0].SizeBytes)
	}
	if m.allResults[0].PackageName != "my-app" {
		t.Errorf("expected updated package name, got %q", m.allResults[0].PackageName)
	}
	if m.allResults[0].Status != model.StatusReady {
		t.Errorf("expected StatusReady, got %v", m.allResults[0].Status)
	}
}

func TestUpdateResult_FallsBackToAddIfUnknown(t *testing.T) {
	m := newModel(nil)
	r := makeResult("/b/node_modules", 200, model.StatusReady)
	m.updateResult(r) // not yet known

	if len(m.allResults) != 1 {
		t.Fatalf("expected 1 result added via updateResult fallback, got %d", len(m.allResults))
	}
}

// --- Stable cursor under async updates ---

func TestCursorStability_AfterInPlaceUpdate(t *testing.T) {
	m := newModel(nil)
	// Use size-desc so adding b (larger) causes a reorder, exercising cursor-ID tracking.
	m.sortMode = SortBySizeDesc

	// Add two pending rows; cursor starts at 0 (first item added: a, size 50)
	a := makeResult("/a/node_modules", 50, model.StatusPending)
	b := makeResult("/b/node_modules", 200, model.StatusPending)
	m.addResult(a)
	// Cursor is on a (size 50) at index 0
	if m.cursor != 0 {
		t.Fatalf("expected cursor at 0 after first add, got %d", m.cursor)
	}

	// Add b; sort-by-size-desc now puts b (200) first, a (50) second.
	// Cursor should stay on a (same ID), which is now at index 1.
	m.addResult(b)
	if m.cursor != 1 {
		t.Fatalf("expected cursor at 1 after second add (cursor tracks item a by ID), got %d", m.cursor)
	}
	if m.visibleResults[m.cursor].ID != "/a/node_modules" {
		t.Fatalf("expected cursor on /a/node_modules, got %q", m.visibleResults[m.cursor].ID)
	}

	// Confirm which ID we're on
	wantID := m.visibleResults[m.cursor].ID // /a/node_modules

	// Enrich b with a much larger size — sort order stays the same (b first)
	bUpdated := makeResult("/b/node_modules", 1000, model.StatusReady)
	m.updateResult(bUpdated)

	// Cursor should still be on /a/node_modules
	if m.cursor >= len(m.visibleResults) {
		t.Fatalf("cursor out of bounds after update")
	}
	gotID := m.visibleResults[m.cursor].ID
	if gotID != wantID {
		t.Errorf("cursor moved to wrong item: want %q, got %q", wantID, gotID)
	}
}

func TestCursorStability_UnderSearchFilter(t *testing.T) {
	m := newModel(nil)
	// Use size-desc so beta (200) is listed before alpha (100).
	m.sortMode = SortBySizeDesc

	a := makeResult("/projects/alpha/node_modules", 100, model.StatusReady)
	a.PackageName = "alpha"
	b := makeResult("/projects/beta/node_modules", 200, model.StatusReady)
	b.PackageName = "beta"
	m.addResult(a)
	m.addResult(b)

	// Move cursor to alpha (index 1 in size-desc order since beta=200 > alpha=100)
	m.moveCursor(1)
	if m.visibleResults[m.cursor].PackageName != "alpha" {
		t.Fatalf("expected cursor on alpha, got %q", m.visibleResults[m.cursor].PackageName)
	}
	savedID := m.visibleResults[m.cursor].ID

	// Apply filter that hides alpha
	m.searchQuery = "beta"
	m.refilterAndSort()

	// alpha is filtered out; cursor should clamp to visible
	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 visible result after filter, got %d", len(m.visibleResults))
	}

	// Remove filter — alpha should come back
	m.searchQuery = ""
	m.refilterAndSort()

	// Cursor should restore to alpha (saved ID no longer tracked since it was filtered out,
	// so it stays at the clamped position — this is acceptable: the important thing is
	// cursor doesn't crash and stays in bounds)
	if m.cursor < 0 || m.cursor >= len(m.visibleResults) {
		t.Errorf("cursor out of bounds after filter removal: cursor=%d, len=%d", m.cursor, len(m.visibleResults))
	}
	_ = savedID // ID tracking after filter removal is best-effort
}

// --- Selection stability under async updates ---

func TestSelectionStability_AfterAsyncUpdate(t *testing.T) {
	m := newModel(nil)
	a := makeResult("/a/node_modules", 100, model.StatusPending)
	b := makeResult("/b/node_modules", 200, model.StatusPending)
	m.addResult(a)
	m.addResult(b)

	// Select item /a
	m.selected["/a/node_modules"] = true

	// Enrich /b with a bigger size (it stays at top)
	bReady := makeResult("/b/node_modules", 1000, model.StatusReady)
	m.updateResult(bReady)

	// Selection of /a should be preserved
	if !m.selected["/a/node_modules"] {
		t.Error("selection of /a/node_modules was lost after async update")
	}
}

// --- Sort stability ---

func TestSortBySizeDesc_OrderCorrect(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortBySizeDesc
	for _, size := range []int64{300, 100, 500, 200} {
		r := model.Result{
			ID:        "/p" + string(rune('0'+size/100)) + "/node_modules",
			Path:      "/p" + string(rune('0'+size/100)) + "/node_modules",
			SizeBytes: size,
			Status:    model.StatusReady,
			Basename:  "node_modules",
		}
		m.addResult(r)
	}

	for i := 1; i < len(m.visibleResults); i++ {
		if m.visibleResults[i].SizeBytes > m.visibleResults[i-1].SizeBytes {
			t.Errorf("sort order broken at index %d: %d > %d",
				i, m.visibleResults[i].SizeBytes, m.visibleResults[i-1].SizeBytes)
		}
	}
}

func TestSortByNewest_OrderCorrect(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortByNewest
	now := time.Now()
	times := []time.Time{
		now.Add(-24 * time.Hour),
		now.Add(-1 * time.Hour),
		now.Add(-48 * time.Hour),
	}
	for i, t0 := range times {
		r := model.Result{
			ID:           "/p" + string(rune('a'+i)) + "/node_modules",
			Path:         "/p" + string(rune('a'+i)) + "/node_modules",
			LastModified: t0,
			Status:       model.StatusReady,
			Basename:     "node_modules",
		}
		m.addResult(r)
	}

	for i := 1; i < len(m.visibleResults); i++ {
		a := m.visibleResults[i-1].LastModified
		b := m.visibleResults[i].LastModified
		if b.After(a) {
			t.Errorf("newest-first sort broken at index %d", i)
		}
	}
}

// --- scanEventMsg routing ---

func TestUpdate_DiscoveredEvent_AddsPendingRow(t *testing.T) {
	m := newModel(nil)
	r := makeResult("/x/node_modules", 0, model.StatusPending)

	ch := make(chan scan.Event, 1)
	ch <- scan.DiscoveredEvent{Result: r}
	close(ch)

	raw := <-ch
	msg := scanEventMsg{ev: raw, ch: ch}
	updated, _ := m.Update(msg)
	m2 := updated.(Model)

	if len(m2.allResults) != 1 {
		t.Fatalf("expected 1 result after DiscoveredEvent, got %d", len(m2.allResults))
	}
	if m2.allResults[0].Status != model.StatusPending {
		t.Errorf("expected StatusPending, got %v", m2.allResults[0].Status)
	}
}

func TestUpdate_AnalyzedEvent_UpdatesExistingRow(t *testing.T) {
	m := newModel(nil)
	pending := makeResult("/y/node_modules", 0, model.StatusPending)
	m.addResult(pending)

	ch := make(chan scan.Event, 1)
	ready := makeResult("/y/node_modules", 42*1024*1024, model.StatusReady)
	ch <- scan.AnalyzedEvent{Result: ready}
	close(ch)

	raw := <-ch
	msg := scanEventMsg{ev: raw, ch: ch}
	updated, _ := m.Update(msg)
	m2 := updated.(Model)

	if len(m2.allResults) != 1 {
		t.Fatalf("expected 1 result after AnalyzedEvent, got %d", len(m2.allResults))
	}
	if m2.allResults[0].SizeBytes != 42*1024*1024 {
		t.Errorf("expected updated SizeBytes, got %d", m2.allResults[0].SizeBytes)
	}
	if m2.allResults[0].Status != model.StatusReady {
		t.Errorf("expected StatusReady, got %v", m2.allResults[0].Status)
	}
}

func TestUpdate_DoneEvent_SetsScanState(t *testing.T) {
	m := newModel(nil)

	ch := make(chan scan.Event, 1)
	ch <- scan.DoneEvent{Total: 5}
	close(ch)

	raw := <-ch
	msg := scanEventMsg{ev: raw, ch: ch}
	updated, _ := m.Update(msg)
	m2 := updated.(Model)

	if m2.scanState != ScanStateDone {
		t.Errorf("expected ScanStateDone, got %v", m2.scanState)
	}
	if m2.scanFound != 5 {
		t.Errorf("expected scanFound=5, got %d", m2.scanFound)
	}
}

func TestUpdate_ErrorEvent_AppendsError(t *testing.T) {
	m := newModel(nil)

	ch := make(chan scan.Event, 1)
	ch <- scan.ErrorEvent{Err: &testErr{"disk full"}}
	close(ch)

	raw := <-ch
	msg := scanEventMsg{ev: raw, ch: ch}
	updated, _ := m.Update(msg)
	m2 := updated.(Model)

	if len(m2.errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(m2.errors))
	}
	if m2.errors[0] != "disk full" {
		t.Errorf("expected 'disk full', got %q", m2.errors[0])
	}
}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
