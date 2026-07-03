package coverage_test

import (
	"strings"
	"testing"

	"nocrap/internal/coverage"
)

func TestParsePythonCoverage(t *testing.T) {
	cov, err := coverage.ParsePythonCoverage("testdata/python/.coverage.json")
	if err != nil {
		t.Fatalf("ParsePythonCoverage: %v", err)
	}
	key := "testdata/python/simple.py"
	data, ok := cov[key]
	if !ok {
		t.Fatalf("expected key %q in coverage map, got keys: %v", key, mapKeys(cov))
	}
	if !data.CoveredLines[1] {
		t.Error("line 1 should be covered")
	}
	if data.CoveredLines[6] {
		t.Error("line 6 should NOT be covered")
	}
}

func TestParseLCOV(t *testing.T) {
	cov, err := coverage.ParseLCOV("testdata/javascript/lcov.info")
	if err != nil {
		t.Fatalf("ParseLCOV: %v", err)
	}
	key := "testdata/javascript/simple.js"
	data, ok := cov[key]
	if !ok {
		t.Fatalf("expected key %q in coverage map", key)
	}
	if !data.CoveredLines[1] {
		t.Error("line 1 should be covered")
	}
	if data.CoveredLines[3] {
		t.Error("line 3 should NOT be covered (count=0)")
	}
	if data.TotalLines != 7 {
		t.Errorf("TotalLines = %d, want 7", data.TotalLines)
	}
}

func TestParseGoCover(t *testing.T) {
	cov, err := coverage.ParseGoCover("testdata/go/cover.out")
	if err != nil {
		t.Fatalf("ParseGoCover: %v", err)
	}
	key := "calculator.go"
	data, ok := cov[key]
	if !ok {
		for k := range cov {
			if strings.Contains(k, "calculator.go") {
				data = cov[k]
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("expected calculator.go in coverage map, got keys: %v", mapKeys(cov))
		}
	}
	if !data.CoveredLines[10] {
		t.Error("lines 10-13 should be covered (count=1)")
	}
	if data.CoveredLines[14] {
		t.Error("line 14 should NOT be covered (count=0)")
	}
}

func mapKeys(m coverage.CoverageMap) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
