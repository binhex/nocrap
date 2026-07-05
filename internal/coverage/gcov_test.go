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

func TestParseGcov_MultipleLines(t *testing.T) {
	tmpDir := t.TempDir()
	gcovPath := filepath.Join(tmpDir, "multi.gcov")

	content := `        -:    0:Source:/path/to/multi.c
        1:    1:int a() { return 1; }
        1:    2:int b() { return 2; }
    #####:    3:int c() { return 3; }
        1:    4:int d() { return 4; }
        0:    5:int e() { return 5; }
        -:    6:}
`
	if err := os.WriteFile(gcovPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cov, err := coverage.ParseGcov(gcovPath)
	if err != nil {
		t.Fatalf("ParseGcov: %v", err)
	}

	data, ok := cov["/path/to/multi.c"]
	if !ok {
		t.Fatal("missing coverage key")
	}

	if data.TotalLines != 5 {
		t.Errorf("TotalLines = %d, want 5", data.TotalLines)
	}
	if !data.CoveredLines[1] {
		t.Error("line 1 should be covered")
	}
	if !data.CoveredLines[2] {
		t.Error("line 2 should be covered")
	}
	if data.CoveredLines[3] {
		t.Error("line 3 should NOT be covered (#####)")
	}
	if !data.CoveredLines[4] {
		t.Error("line 4 should be covered")
	}
	if data.CoveredLines[5] {
		t.Error("line 5 should NOT be covered (count=0)")
	}
}
