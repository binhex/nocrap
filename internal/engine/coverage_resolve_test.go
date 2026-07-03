package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/config"
	"nocrap/internal/engine"
)

func TestCoverage_ResolvedFromSourceDirParent(t *testing.T) {
	// Simulate the scenario: user runs nocrap from /data/nocrap with:
	//   $ ./nocrap ../boozarr/src
	// where ../boozarr/coverage.json exists but ../boozarr/src/coverage.json does not.
	//
	// Directory structure:
	//   tmpdir/
	//     coverage.json       <-- coverage data (parent of source dir)
	//     source/
	//       lib.py            <-- source file

	tmpDir := t.TempDir()

	// Create source/lib.py
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(srcDir, "lib.py")
	srcContent := []byte(`def hello():
    return "world"

def greet(name):
    if name:
        return f"Hello {name}"
    return "Hello"
`)
	if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create coverage.json at tmpDir root (parent of source/),
	// using the key "source/lib.py" matching the WalkDir path.
	covContent := `{
    "meta": {"format": 2},
    "files": {
        "source/lib.py": {
            "executed_lines": [1, 2, 6, 7],
            "summary": {"covered_lines": 4, "num_statements": 4, "percent_covered": 100},
            "missing_lines": [],
            "excluded_lines": []
        }
    },
    "totals": {
        "covered_lines": 4,
        "num_statements": 4,
        "percent_covered": 100
    }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "coverage.json"), []byte(covContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Use default config — coverage path defaults to "coverage.json" (relative)
	cfg := config.DefaultConfig()

	// Run Analyze with source=tmpDir/source.
	// If resolution works, it should find coverage.json in tmpDir/ (parent of source/).
	scores, err := engine.Analyze([]string{srcDir}, cfg)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(scores) == 0 {
		t.Fatal("expected at least 1 function score")
	}

	for _, s := range scores {
		t.Logf("%s: CC=%d, Cov=%.1f%%, CRAP=%.2f", s.Name, s.CC, s.CoveragePercent, s.CRAP)
		if s.CoveragePercent <= 0 {
			t.Errorf("coverage should be > 0 for %s (got %.1f%%) — coverage file not resolved", s.Name, s.CoveragePercent)
		}
	}
}

func TestCoverage_ResolvedFromSourceDirItself(t *testing.T) {
	// Coverage file is INSIDE the source directory itself.
	// Directory structure:
	//   tmpdir/
	//     project/
	//       src/
	//         lib.py
	//         .coverage.json

	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "project", "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(srcDir, "lib.py")
	srcContent := []byte(`def add(a, b):
    return a + b
`)
	if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Put .coverage.json in the source dir itself
	covContent := `{
    "meta": {"format": 2},
    "files": {
        "lib.py": {
            "executed_lines": [1],
            "summary": {"covered_lines": 1, "num_statements": 1, "percent_covered": 100},
            "missing_lines": [],
            "excluded_lines": []
        }
    },
    "totals": {
        "covered_lines": 1,
        "num_statements": 1,
        "percent_covered": 100
    }
}`
	covFile := filepath.Join(srcDir, ".coverage.json")
	if err := os.WriteFile(covFile, []byte(covContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Coverage.Python = ".coverage.json"

	scores, err := engine.Analyze([]string{srcDir}, cfg)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(scores) == 0 {
		t.Fatal("expected at least 1 function score")
	}

	for _, s := range scores {
		t.Logf("%s: CC=%d, Cov=%.1f%%, CRAP=%.2f", s.Name, s.CC, s.CoveragePercent, s.CRAP)
		if s.CoveragePercent <= 0 {
			t.Errorf("coverage should be > 0 for %s — coverage file not resolved from source dir", s.Name)
		}
	}
}
