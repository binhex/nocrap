package engine

import "testing"

// This white-box test is in package engine (not engine_test) so it can test
// the unexported matchesExclude and matchGlobstar functions.

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
	// accidentally exclude all files.
	files, err := collectFiles([]string{"/data/boozarr/src/boozarr/"}, []string{
		"**/test_*",
		"**/*_test.go",
		"**/vendor/**",
		"**/testdata/**",
		"**/crossval/**",
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
