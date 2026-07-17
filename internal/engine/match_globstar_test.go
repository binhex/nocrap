package engine

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/coverage"
)

// This white-box test is in package engine (not engine_test) so it can test
// the unexported matchesExclude, matchGlobstar, and computeCoverage functions.

func TestMatchesExclude(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		// Simple filepath.Match patterns
		{"literal match", "cli.py", "cli.py", true},
		{"literal no match", "cli.py", "db.py", false},
		{"glob star", "*.py", "cli.py", true},
		{"glob star no match", "*.go", "cli.py", false},

		// **/test_* — matches basename starting with "test_" anywhere
		{"globstar test_ matches", "**/test_*", "/data/project/test_foo.py", true},
		{"globstar test_ not in path", "**/test_*", "cli.py", false},
		{"globstar test_ in subdir", "**/test_*", "/data/project/subdir/test_utils.py", true},

		// **/*_test.go — matches Go test files anywhere
		{"globstar test.go matches", "**/*_test.go", "/data/project/pkg/file_test.go", true},
		{"globstar test.go no match py", "**/*_test.go", "/data/project/cli.py", false},

		// **/vendor/** — matches files inside vendor directories (REGARDLESS of nesting)
		{"globstar vendor nested", "**/vendor/**", "/data/project/vendor/pkg/file.go", true},
		{"globstar vendor deep", "**/vendor/**", "/data/project/src/vendor/lib/file.js", true},
		{"globstar vendor no match", "**/vendor/**", "/data/boozarr/src/cli.py", false},

		// **/testdata/** — matches files inside testdata directories
		{"globstar testdata matches", "**/testdata/**", "/data/project/testdata/python/simple.py", true},
		{"globstar testdata no match", "**/testdata/**", "/data/boozarr/src/cli.py", false},

		// **/crossval/** — matches files inside crossval directories
		{"globstar crossval matches", "**/crossval/**", "/data/nocrap/crossval/crossval_test.go", true},
		{"globstar crossval no match", "**/crossval/**", "/data/boozarr/src/cli.py", false},

		// Prefix-only patterns: src/** matches paths starting with src/
		{"prefix star matches", "src/**", "src/cli.py", true},
		{"prefix star subdir", "src/**", "src/subdir/util.go", true},
		{"prefix star no match", "src/**", "/data/src/cli.py", false},

		// Multi-segment ** patterns
		{"multi ** two segments", "**/a/**/b/**", "/data/project/a/1/b/file.go", true},
		{"multi ** two segments wrong order", "**/b/**/a/**", "/data/project/a/1/b/file.go", false},
		{"multi ** with preceding text", "**/boozarr/**/src/**", "/data/boozarr/src/cli.py", true},

		// Empty pattern
		{"empty pattern", "", "cli.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesExclude(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchesExclude(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchGlobstar_CorpusExcludes(t *testing.T) {
	// Verify the exact exclude patterns from .crap.toml against boozarr paths.
	excludes := []string{
		"**/test_*",
		"**/*_test.go",
		"**/vendor/**",
		"**/testdata/**",
		"**/crossval/**",
	}
	boozarrFiles := []string{
		"/data/boozarr/src/boozarr/cli.py",
		"/data/boozarr/src/boozarr/db.py",
		"/data/boozarr/src/boozarr/epub.py",
		"/data/boozarr/src/boozarr/logger.py",
		"/data/boozarr/src/boozarr/pipeline.py",
		"/data/boozarr/src/boozarr/report.py",
		"/data/boozarr/src/boozarr/utils.py",
	}
	shouldBeExcluded := []string{
		"/data/project/testdata/fixtures/input.json",
		"/data/project/vendor/lib/bundle.js",
		"/data/project/crossval/corpus/test.py",
	}

	for _, pattern := range excludes {
		for _, path := range boozarrFiles {
			if matchesExclude(pattern, path) {
				t.Errorf("%s should NOT match %q (boozarr source file)", pattern, path)
			}
		}
		_ = shouldBeExcluded // used for verification that patterns DO match intended files
	}

	// Verify patterns DO match the files they should exclude
	expectMatch := map[string]string{
		"**/test_*":      "/data/project/tests/test_foo.py",
		"**/*_test.go":   "/data/project/pkg/file_test.go",
		"**/vendor/**":   "/data/project/vendor/lib/pkg.js",
		"**/testdata/**": "/data/project/testdata/fixtures/input.json",
		"**/crossval/**": "/data/project/crossval/crossval_test.go",
	}
	for pattern, path := range expectMatch {
		if !matchesExclude(pattern, path) {
			t.Errorf("%s SHOULD match %q but didn't", pattern, path)
		}
	}
}

func TestCollectFilesWithExcludes(t *testing.T) {
	// Integration test: verify collectFiles with real exclude patterns doesn't
	// accidentally exclude all files. Uses a temp dir so it works in CI.
	tmpDir := t.TempDir()
	includeDirs := []string{
		filepath.Join(tmpDir, "src/boozarr"),
		filepath.Join(tmpDir, "src/boozarr/utils"),
	}
	for _, d := range includeDirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	includeFiles := []string{
		filepath.Join(tmpDir, "src/boozarr/cli.py"),
		filepath.Join(tmpDir, "src/boozarr/config.py"),
		filepath.Join(tmpDir, "src/boozarr/utils/helpers.py"),
	}
	excludeFiles := []string{
		filepath.Join(tmpDir, "src/boozarr/tests/test_cli.py"),
		filepath.Join(tmpDir, "src/boozarr/vendor/lib/dep.py"),
		filepath.Join(tmpDir, "src/boozarr/testdata/fixtures/data.py"),
	}
	for _, f := range append(includeFiles, excludeFiles...) {
		dir := filepath.Dir(f)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.WriteFile(f, []byte("x = 1\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}

	files, err := collectFiles([]string{filepath.Join(tmpDir, "src/boozarr/")}, []string{
		"**/tests/**",
		"**/vendor/**",
		"**/testdata/**",
	})
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("collectFiles returned 0 files — exclude patterns are over-matching")
	}

	// We expect at least some Python files from boozarr
	countPy := 0
	for _, f := range files {
		if len(f) > 3 && f[len(f)-3:] == ".py" {
			countPy++
		}
	}
	if countPy == 0 {
		t.Errorf("expected at least 1 .py file, got %d total files: %v", len(files), files)
	}
}

func TestComputeCoverage_RelativePathInCoverageMap(t *testing.T) {
	source := []byte("def add(a, b):\n    return a + b\n")
	source2 := []byte("x = 1\ny = 2\nz = 3\n")

	covMap := coverage.CoverageMap{
		"src/boozarr/cli.py": &coverage.CoverageData{
			CoveredLines: map[int]bool{1: true, 2: true},
			TotalLines:   3,
		},
		"other/unrelated.py": &coverage.CoverageData{
			CoveredLines: map[int]bool{1: true, 2: true, 3: true},
			TotalLines:   3,
		},
	}

	// Absolute path from WalkDir — should match "src/boozarr/cli.py" via suffix
	got := computeCoverage(covMap, "/data/boozarr/src/boozarr/cli.py", 1, 2, source)
	if got <= 0.0 {
		t.Errorf("absolute path matching relative key: got %.1f%%, want > 0", got)
	}

	// Exact match
	got = computeCoverage(covMap, "src/boozarr/cli.py", 1, 2, source)
	if got <= 0.0 {
		t.Errorf("exact relative path: got %.1f%%, want > 0", got)
	}

	// Basename fallback
	covMapBase := coverage.CoverageMap{
		"cli.py": &coverage.CoverageData{
			CoveredLines: map[int]bool{1: true, 2: true},
			TotalLines:   3,
		},
	}
	got = computeCoverage(covMapBase, "/any/path/cli.py", 1, 2, source)
	if got <= 0.0 {
		t.Errorf("basename fallback: got %.1f%%, want > 0", got)
	}

	// No match
	got = computeCoverage(covMap, "/nonexistent/path.py", 1, 2, source2)
	if got > 0.0 {
		t.Errorf("no matching key: got %.1f%%, want 0.0", got)
	}

	// False positive: covMap has "cli.py", filePath has "old_cli.py"
	// The basename fallback already handles bare-filename keys correctly;
	// suffix matching must NOT match partial filenames without path boundary.
	covMapPartial := coverage.CoverageMap{
		"cli.py": &coverage.CoverageData{
			CoveredLines: map[int]bool{1: true, 2: true},
			TotalLines:   3,
		},
	}
	got = computeCoverage(covMapPartial, "/project/old_cli.py", 1, 2, source)
	if got > 0.0 {
		t.Errorf("false positive: old_cli.py should not match cli.py, got %.1f%%", got)
	}
	// Same key via basename should still work
	got = computeCoverage(covMapPartial, "/any/path/cli.py", 1, 2, source)
	if got <= 0.0 {
		t.Errorf("basename cli.py should match, got %.1f%%", got)
	}
}
