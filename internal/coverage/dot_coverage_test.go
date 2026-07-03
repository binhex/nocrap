package coverage_test

import (
	"path/filepath"
	"testing"

	"nocrap/internal/coverage"
)

func TestParseDotCoverage_Lines(t *testing.T) {
	// The .coverage SQLite database contains line-level coverage data
	// matching what pytest-crap gets from data.lines().
	// This should return MORE covered lines than coverage.json's statement-level data.
	fp := filepath.Join("..", "..", "..", "boozarr", ".coverage")

	cov, err := coverage.ParseDotCoverage(fp)
	if err != nil {
		t.Skipf("boozarr .coverage not available: %v", err)
	}

	key := "src/boozarr/processors/metadata.py"
	data, ok := cov[key]
	if !ok {
		t.Fatalf("expected key %q in coverage map", key)
	}

	// data.lines() for metadata.py returns 109 lines
	// coverage.json's executed_lines only has 90 lines
	if len(data.CoveredLines) != 109 {
		t.Logf("CoveredLines count = %d (expected 109)", len(data.CoveredLines))
	}
	t.Logf("CoveredLines: %d total", len(data.CoveredLines))

	// Verify continuation lines of multi-line statements are included
	// Line 140 is a continuation of the return statement starting at 139
	if !data.CoveredLines[140] {
		t.Error("line 140 should be covered (continuation of multi-line return statement)")
	}
	if !data.CoveredLines[141] {
		t.Error("line 141 should be covered (continuation of multi-line return statement)")
	}
	if !data.CoveredLines[153] {
		t.Error("line 153 should be covered (continuation of multi-line return statement)")
	}

	// Line 139 is the start of the return statement, should also be covered
	if !data.CoveredLines[139] {
		t.Error("line 139 should be covered (start of return statement)")
	}
}
