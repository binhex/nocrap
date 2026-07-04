package reporter

import (
	"bytes"
	"strings"
	"testing"

	"nocrap/internal/engine"
)

func makeScores() []engine.FunctionScore {
	return []engine.FunctionScore{
		{Name: "bigFunc", File: "/tmp/src/pkg/a.py", StartLine: 1, EndLine: 50, CC: 20, CoveragePercent: 0, CRAP: 420},
		{Name: "medFunc", File: "/tmp/src/pkg/a.py", StartLine: 51, EndLine: 80, CC: 10, CoveragePercent: 50, CRAP: 22.5},
		{Name: "smallFunc", File: "/tmp/src/pkg/b.py", StartLine: 1, EndLine: 10, CC: 4, CoveragePercent: 80, CRAP: 4.32},
		{Name: "cleanFunc", File: "/tmp/src/pkg/b.py", StartLine: 11, EndLine: 20, CC: 2, CoveragePercent: 100, CRAP: 2},
		{Name: "otherPkgFunc", File: "/tmp/src/sub/c.py", StartLine: 1, EndLine: 30, CC: 8, CoveragePercent: 90, CRAP: 8.06},
	}
}

func TestNew_DefaultWidth(t *testing.T) {
	r := New("")
	if r.width < 1 || r.width > 200 {
		t.Errorf("expected width 1-200, got %d", r.width)
	}
	if r.rootDir != "" {
		t.Errorf("expected empty rootDir, got %q", r.rootDir)
	}
}

func TestNew_WithRootDir(t *testing.T) {
	r := New("/my/project")
	if r.rootDir != "/my/project" {
		t.Errorf("expected '/my/project', got %q", r.rootDir)
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		crap  float64
		color string
	}{
		{50, colorRed},
		{31, colorRed},
		{30, colorYellow},
		{16, colorYellow},
		{15, colorGreen},
		{9, colorGreen},
		{0, colorGreen},
	}
	for _, tc := range tests {
		got := colorize(tc.crap)
		if got != tc.color {
			t.Errorf("colorize(%.0f) = %q, want %q", tc.crap, got, tc.color)
		}
	}
}

func TestRelativePath(t *testing.T) {
	r := New("/tmp/src")
	got := r.relativePath("/tmp/src/pkg/a.py")
	if got != "pkg/a.py" {
		t.Errorf("expected 'pkg/a.py', got %q", got)
	}
}

func TestRelativePath_EmptyRootDir(t *testing.T) {
	r := New("")
	got := r.relativePath("/foo/bar/baz.py")
	if got != "/foo/bar/baz.py" {
		t.Errorf("expected unchanged path, got %q", got)
	}
}

func TestTruncateMiddle(t *testing.T) {
	tests := []struct {
		s, want string
		maxLen  int
	}{
		{"short", "short", 20},
		{"1234567890abcdef", "12345...bcdef", 13},
		{"abcd", "a", 1},
	}
	for _, tc := range tests {
		got := truncateMiddle(tc.s, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncateMiddle(%q, %d) = %q, want %q", tc.s, tc.maxLen, got, tc.want)
		}
	}
}

func TestTruncateRight(t *testing.T) {
	tests := []struct {
		s, want string
		maxLen  int
	}{
		{"short", "short", 20},
		{"1234567890abcdef", "12345...", 8},
		{"abc", "a", 1},
	}
	for _, tc := range tests {
		got := truncateRight(tc.s, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncateRight(%q, %d) = %q, want %q", tc.s, tc.maxLen, got, tc.want)
		}
	}
}

func TestGroupScores(t *testing.T) {
	scores := makeScores()
	groups := groupScores(scores, 9, func(s engine.FunctionScore) string { return s.File })
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	// groupScores must be sorted by maxCRAP descending
	if groups[0].maxCRAP < groups[1].maxCRAP {
		t.Error("groups not sorted descending by maxCRAP")
	}
	// "/tmp/src/pkg/a.py" has bigFunc (420), medFunc (22.5) → max=420, countAbove=2
	// "/tmp/src/pkg/b.py" has smallFunc (4.32), cleanFunc (2) → max=4.32, countAbove=0
	var aGroup *fileSummary
	for _, g := range groups {
		if g.file == "/tmp/src/pkg/a.py" {
			aGroup = g
		}
	}
	if aGroup == nil {
		t.Fatal("missing group for /tmp/src/pkg/a.py")
	}
	if aGroup.maxCRAP != 420 {
		t.Errorf("expected maxCRAP=420, got %.2f", aGroup.maxCRAP)
	}
	if aGroup.countAbove != 2 {
		t.Errorf("expected countAbove=2, got %d", aGroup.countAbove)
	}
}

func TestRenderFunctionTable(t *testing.T) {
	r := New("/tmp/src")
	scores := makeScores()
	var buf bytes.Buffer
	r.renderFunctionTable(&buf, scores, 0)
	out := buf.String()
	if !strings.Contains(out, "CRAP by Function") {
		t.Error("missing table header")
	}
	if !strings.Contains(out, "bigFunc") {
		t.Error("missing bigFunc")
	}
}

func TestRenderFunctionTable_TopN(t *testing.T) {
	r := New("/tmp/src")
	var buf bytes.Buffer
	r.renderFunctionTable(&buf, makeScores(), 2)
	out := buf.String()
	if strings.Count(out, "\n") > 7 {
		t.Error("expected only 2 rows (header + 2 data lines)")
	}
}

func TestRenderGrouped_File(t *testing.T) {
	r := New("/tmp/src")
	scores := makeScores()
	groups := groupScores(scores, 9, func(s engine.FunctionScore) string { return s.File })
	var buf bytes.Buffer
	renderGrouped(&buf, "File", groups, 0, r)
	out := buf.String()
	if !strings.Contains(out, "CRAP by File") {
		t.Error("missing grouped header")
	}
}

func TestRenderFileSummary(t *testing.T) {
	r := New("/tmp/src")
	// Capture os.Stdout is tricky, just verify no panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderFileSummary panicked: %v", r)
		}
	}()
	r.RenderFileSummary(makeScores(), 0, 9)
}

func TestRenderFolderSummary(t *testing.T) {
	r := New("/tmp/src")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderFolderSummary panicked: %v", r)
		}
	}()
	r.RenderFolderSummary(makeScores(), 0, 9)
}

func TestRenderGrouped_TopN(t *testing.T) {
	r := New("/tmp/src")
	groups := groupScores(makeScores(), 0, func(s engine.FunctionScore) string { return s.File })
	var buf bytes.Buffer
	renderGrouped(&buf, "File", groups, 2, r)
	out := buf.String()
	lines := strings.Count(out, "\n")
	if lines > 6 {
		t.Errorf("expected at most 6 lines with topN=2, got %d", lines)
	}
}
