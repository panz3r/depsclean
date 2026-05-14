package tui

import (
	"testing"
	"time"

	"github.com/panz3r/depsclean/internal/model"
)

// --- Default sort mode tests ---

// TestDefaultSortMode_IsPathAsc verifies that a freshly created model uses
// path-ascending order so related nested projects stay adjacent.
func TestDefaultSortMode_IsPathAsc(t *testing.T) {
	m := newModel(nil)
	if m.sortMode != SortByPathAsc {
		t.Errorf("expected default sortMode to be SortByPathAsc (%d), got %d", SortByPathAsc, m.sortMode)
	}
}

// TestDefaultSort_NestedProjectsAdjacent checks that projects sharing a common
// root directory are listed consecutively under the default path sort.
func TestDefaultSort_NestedProjectsAdjacent(t *testing.T) {
	m := newModel(nil)
	// Three siblings under /work/mono and one unrelated project.
	paths := []struct {
		id, projectPath string
	}{
		{"/work/mono/app/node_modules", "/work/mono/app"},
		{"/other/standalone/node_modules", "/other/standalone"},
		{"/work/mono/api/node_modules", "/work/mono/api"},
		{"/work/mono/shared/node_modules", "/work/mono/shared"},
	}
	for i, p := range paths {
		r := model.Result{
			ID:          p.id,
			Path:        p.id,
			ProjectPath: p.projectPath,
			Basename:    "node_modules",
			SizeBytes:   int64((i + 1) * 100),
			Status:      model.StatusReady,
		}
		m.addResult(r)
	}

	// All /work/mono/* entries must be contiguous in the sorted output.
	monoStart := -1
	monoEnd := -1
	for i, r := range m.visibleResults {
		if len(r.ProjectPath) >= len("/work/mono") && r.ProjectPath[:len("/work/mono")] == "/work/mono" {
			if monoStart == -1 {
				monoStart = i
			}
			monoEnd = i
		} else if monoStart != -1 && monoEnd != -1 && i > monoEnd {
			// Non-mono entry appearing after mono entries started — OK only if all mono are done.
			// Check if any more mono entries appear after this.
			for j := i + 1; j < len(m.visibleResults); j++ {
				rj := m.visibleResults[j]
				if len(rj.ProjectPath) >= len("/work/mono") && rj.ProjectPath[:len("/work/mono")] == "/work/mono" {
					t.Errorf("mono projects are not adjacent: non-mono entry at index %d breaks the group (next mono at %d)", i, j)
					return
				}
			}
		}
	}
	if monoStart == -1 {
		t.Fatal("no /work/mono entries found in visible results")
	}
}

// TestDefaultSort_PathOrderStability verifies that results with the same
// ProjectPath preserve a stable relative order (i.e., sort.Slice tie-breaking
// does not scramble equal-path entries across repeated refilterAndSort calls).
func TestDefaultSort_PathOrderStability(t *testing.T) {
	m := newModel(nil)
	// Two entries sharing the same project path but different dep directories.
	entries := []model.Result{
		{
			ID:          "/workspace/myapp/node_modules",
			Path:        "/workspace/myapp/node_modules",
			ProjectPath: "/workspace/myapp",
			Basename:    "node_modules",
			SizeBytes:   500,
			Status:      model.StatusReady,
		},
		{
			ID:          "/workspace/myapp/.cache",
			Path:        "/workspace/myapp/.cache",
			ProjectPath: "/workspace/myapp",
			Basename:    ".cache",
			SizeBytes:   200,
			Status:      model.StatusReady,
		},
	}
	for _, r := range entries {
		m.addResult(r)
	}

	// Record the initial order of IDs.
	first := make([]string, len(m.visibleResults))
	for i, r := range m.visibleResults {
		first[i] = r.ID
	}

	// Re-run refilterAndSort multiple times — order must be deterministic.
	for round := 0; round < 5; round++ {
		m.refilterAndSort()
		for i, r := range m.visibleResults {
			if r.ID != first[i] {
				t.Errorf("round %d: order changed at index %d: want %q, got %q",
					round+1, i, first[i], r.ID)
			}
		}
	}
}

// TestNextSortMode_StartsFromPathAsc verifies that cycling from the default
// sort (SortByPathAsc) moves to the next mode without wrapping to itself.
func TestNextSortMode_StartsFromPathAsc(t *testing.T) {
	next := NextSortMode(SortByPathAsc)
	if next == SortByPathAsc {
		t.Errorf("NextSortMode(SortByPathAsc) returned SortByPathAsc — sort cycling is broken")
	}
}

// --- Additional sort mode tests ---

func TestSortBySizeAsc_OrderCorrect(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortBySizeAsc
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
		if m.visibleResults[i].SizeBytes < m.visibleResults[i-1].SizeBytes {
			t.Errorf("size-asc sort broken at index %d: %d < %d",
				i, m.visibleResults[i].SizeBytes, m.visibleResults[i-1].SizeBytes)
		}
	}
}

func TestSortByNameAsc_OrderCorrect(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortByNameAsc

	names := []string{"zebra-app", "alpha-lib", "mango-pkg"}
	for i, name := range names {
		r := model.Result{
			ID:          "/projects/" + name + "/node_modules",
			Path:        "/projects/" + name + "/node_modules",
			ProjectPath: "/projects/" + name,
			Basename:    "node_modules",
			PackageName: name,
			SizeBytes:   int64((i + 1) * 100),
			Status:      model.StatusReady,
		}
		m.addResult(r)
	}

	for i := 1; i < len(m.visibleResults); i++ {
		a := m.visibleResults[i-1].PackageName
		b := m.visibleResults[i].PackageName
		if a > b {
			t.Errorf("name-asc sort broken at index %d: %q > %q", i, a, b)
		}
	}
}

func TestSortByPathAsc_OrderCorrect(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortByPathAsc

	paths := []string{"/zoo/node_modules", "/alpha/node_modules", "/middle/node_modules"}
	for i, path := range paths {
		r := model.Result{
			ID:          path,
			Path:        path,
			ProjectPath: path + "/..",
			Basename:    "node_modules",
			SizeBytes:   int64((i + 1) * 100),
			Status:      model.StatusReady,
		}
		m.addResult(r)
	}

	for i := 1; i < len(m.visibleResults); i++ {
		a := m.visibleResults[i-1].ProjectPath
		b := m.visibleResults[i].ProjectPath
		if compareHierarchicalPath(a, b) > 0 {
			t.Errorf("path-asc sort broken at index %d: %q should not come before %q", i, a, b)
		}
	}
}

func TestSortByPathAsc_ParentBeforeDescendant(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortByPathAsc

	results := []model.Result{
		{
			ID:          "/repo/packages/api/node_modules",
			Path:        "/repo/packages/api/node_modules",
			ProjectPath: "/repo/packages/api",
			Basename:    "node_modules",
			Status:      model.StatusReady,
		},
		{
			ID:          "/repo/node_modules",
			Path:        "/repo/node_modules",
			ProjectPath: "/repo",
			Basename:    "node_modules",
			Status:      model.StatusReady,
		},
		{
			ID:          "/repo/packages/api/subpkg/node_modules",
			Path:        "/repo/packages/api/subpkg/node_modules",
			ProjectPath: "/repo/packages/api/subpkg",
			Basename:    "node_modules",
			Status:      model.StatusReady,
		},
	}
	for _, r := range results {
		m.addResult(r)
	}

	want := []string{
		"/repo",
		"/repo/packages/api",
		"/repo/packages/api/subpkg",
	}
	for i, projectPath := range want {
		if m.visibleResults[i].ProjectPath != projectPath {
			t.Fatalf("visibleResults[%d] = %q, want %q", i, m.visibleResults[i].ProjectPath, projectPath)
		}
	}
}

func TestSortByOldest_OrderCorrect(t *testing.T) {
	m := newModel(nil)
	m.sortMode = SortByOldest
	now := time.Now()
	times := []time.Time{
		now.Add(-1 * time.Hour),
		now.Add(-48 * time.Hour),
		now.Add(-24 * time.Hour),
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
		if b.Before(a) {
			t.Errorf("oldest-first sort broken at index %d: %v is after %v", i, a, b)
		}
	}
}

// --- Search / filter tests ---

func makeResultWithMeta(id, projectPath, pkgName, pkgVersion, basename string, size int64) model.Result {
	return model.Result{
		ID:             id,
		Path:           id,
		ProjectPath:    projectPath,
		Basename:       basename,
		PackageName:    pkgName,
		PackageVersion: pkgVersion,
		SizeBytes:      size,
		Status:         model.StatusReady,
	}
}

func TestSearch_FiltersByProjectPath(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/workspace/frontend/node_modules", "/workspace/frontend", "frontend-app", "1.0.0", "node_modules", 100))
	m.addResult(makeResultWithMeta("/workspace/backend/node_modules", "/workspace/backend", "backend-svc", "2.0.0", "node_modules", 200))

	m.searchQuery = "frontend"
	m.refilterAndSort()

	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 visible result, got %d", len(m.visibleResults))
	}
	if m.visibleResults[0].PackageName != "frontend-app" {
		t.Errorf("expected frontend-app, got %q", m.visibleResults[0].PackageName)
	}
}

func TestSearch_FiltersByPackageName(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/a/node_modules", "/a", "my-special-lib", "1.0.0", "node_modules", 100))
	m.addResult(makeResultWithMeta("/b/node_modules", "/b", "other-pkg", "1.0.0", "node_modules", 200))

	m.searchQuery = "special"
	m.refilterAndSort()

	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 visible result for package name filter, got %d", len(m.visibleResults))
	}
	if m.visibleResults[0].PackageName != "my-special-lib" {
		t.Errorf("expected my-special-lib, got %q", m.visibleResults[0].PackageName)
	}
}

func TestSearch_FiltersByPackageVersion(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/a/node_modules", "/a", "pkg-a", "3.14.0", "node_modules", 100))
	m.addResult(makeResultWithMeta("/b/node_modules", "/b", "pkg-b", "1.0.0", "node_modules", 200))

	m.searchQuery = "3.14"
	m.refilterAndSort()

	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 result for version filter, got %d", len(m.visibleResults))
	}
	if m.visibleResults[0].PackageVersion != "3.14.0" {
		t.Errorf("expected version 3.14.0, got %q", m.visibleResults[0].PackageVersion)
	}
}

func TestSearch_FiltersByBasename(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/a/node_modules", "/a", "", "", "node_modules", 100))
	m.addResult(makeResultWithMeta("/b/vendor", "/b", "", "", "vendor", 200))

	m.searchQuery = "vendor"
	m.refilterAndSort()

	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 result for basename filter, got %d", len(m.visibleResults))
	}
	if m.visibleResults[0].Basename != "vendor" {
		t.Errorf("expected basename=vendor, got %q", m.visibleResults[0].Basename)
	}
}

func TestSearch_NoMatchReturnsEmpty(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/a/node_modules", "/a", "pkg-a", "1.0.0", "node_modules", 100))
	m.addResult(makeResultWithMeta("/b/node_modules", "/b", "pkg-b", "1.0.0", "node_modules", 200))

	m.searchQuery = "zzz_nomatch"
	m.refilterAndSort()

	if len(m.visibleResults) != 0 {
		t.Fatalf("expected 0 visible results for unmatched query, got %d", len(m.visibleResults))
	}
}

func TestSearch_EmptyQueryRestoresAll(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/a/node_modules", "/a", "pkg-a", "1.0.0", "node_modules", 100))
	m.addResult(makeResultWithMeta("/b/node_modules", "/b", "pkg-b", "1.0.0", "node_modules", 200))

	// Filter down to 1 result.
	m.searchQuery = "pkg-a"
	m.refilterAndSort()
	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 result after filter, got %d", len(m.visibleResults))
	}

	// Clear the query — both results should be visible again.
	m.searchQuery = ""
	m.refilterAndSort()
	if len(m.visibleResults) != 2 {
		t.Fatalf("expected 2 results after clearing filter, got %d", len(m.visibleResults))
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	m := newModel(nil)
	m.addResult(makeResultWithMeta("/a/node_modules", "/a", "MyAwesomeLib", "1.0.0", "node_modules", 100))

	m.searchQuery = "myawesome"
	m.refilterAndSort()

	if len(m.visibleResults) != 1 {
		t.Fatalf("expected 1 result for case-insensitive search, got %d", len(m.visibleResults))
	}
}
