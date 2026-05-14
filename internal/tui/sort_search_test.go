package tui

import (
	"testing"
	"time"

	"github.com/panz3r/npclean/internal/model"
)

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
		if a > b {
			t.Errorf("path-asc sort broken at index %d: %q > %q", i, a, b)
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
