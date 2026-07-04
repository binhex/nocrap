package coverage_test

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/coverage"
)

func TestParseGcov(t *testing.T) {
	tmpDir := t.TempDir()
	gcovPath := filepath.Join(tmpDir, "test.c.gcov")

	content := `        -:    0:Source:/path/to/test.c
        -:    1:#include <stdio.h>
        5:    2:int main() {
        5:    3:    if (1) {
    #####:    4:        return -1;
        -:    5:    }
        5:    6:    return 0;
        -:    7:}
`
	if err := os.WriteFile(gcovPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cov, err := coverage.ParseGcov(gcovPath)
	if err != nil {
		t.Fatalf("ParseGcov: %v", err)
	}

	key := "/path/to/test.c"
	data, ok := cov[key]
	if !ok {
		t.Fatalf("expected key %q in coverage map", key)
	}

	if !data.CoveredLines[2] {
		t.Error("line 2 should be covered")
	}
	if !data.CoveredLines[3] {
		t.Error("line 3 should be covered")
	}
	if data.CoveredLines[4] {
		t.Error("line 4 should NOT be covered (#####)")
	}
	if !data.CoveredLines[6] {
		t.Error("line 6 should be covered")
	}
	if data.TotalLines != 4 {
		t.Errorf("TotalLines = %d, want 4", data.TotalLines)
	}
}

func TestParseGcovMissing(t *testing.T) {
	_, err := coverage.ParseGcov("/nonexistent/file.gcov")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
