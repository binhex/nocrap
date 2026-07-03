package reporter_test

import (
	"bytes"
	"strings"
	"testing"

	"nocrap/internal/engine"
	"nocrap/internal/reporter"
)

func TestWriteJSON(t *testing.T) {
	scores := []engine.FunctionScore{
		{Name: "test_func", File: "/path/to/file.py", CC: 5, CoveragePercent: 80.0, CRAP: 5.2},
		{Name: "bad_func", File: "/path/to/bad.py", CC: 15, CoveragePercent: 20.0, CRAP: 42.0},
	}

	var buf bytes.Buffer
	err := reporter.WriteJSON(scores, &buf)
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test_func") {
		t.Error("JSON output missing test_func")
	}
	if !strings.Contains(output, "bad_func") {
		t.Error("JSON output missing bad_func")
	}
	if !strings.Contains(output, "5.2") {
		t.Error("JSON output missing CRAP score")
	}
}
