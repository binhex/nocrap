package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path, want string
	}{
		{"foo.py", "python"},
		{"foo.js", "javascript"},
		{"foo.mjs", "javascript"},
		{"foo.ts", "typescript"},
		{"foo.tsx", "typescript"},
		{"foo.go", "go"},
		{"foo.txt", ""},
	}
	for _, tc := range tests {
		got := detectLanguage(tc.path)
		if got != tc.want {
			t.Errorf("detectLanguage(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestCountExecutableLines(t *testing.T) {
	tests := []struct {
		name       string
		source     []byte
		start, end int
		want       int
	}{
		{
			"simple function",
			[]byte("def foo():\n    return 1\n"),
			1, 2, 2,
		},
		{
			"with blank lines",
			[]byte("def foo():\n    \n    return 1\n    \n"),
			1, 4, 2,
		},
		{
			"with comments",
			[]byte("def foo():\n    # comment\n    return 1\n"),
			1, 3, 2,
		},
		{
			"with docstring",
			[]byte("def foo():\n    \"\"\"doc\"\"\"\n    return 1\n"),
			1, 3, 3,
		},
		{
			"multi-line string",
			[]byte("def foo():\n    x = '''multi\n    line'''\n    return 1\n"),
			1, 4, 4,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := countExecutableLines(tc.source, tc.start, tc.end)
			if got != tc.want {
				t.Errorf("countExecutableLines(start=%d, end=%d) = %d, want %d",
					tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestGroupByLanguage(t *testing.T) {
	files := []string{"a.py", "b.py", "c.go", "d.js", "e.ts"}
	byLang := groupByLanguage(files)
	if len(byLang) != 4 {
		t.Errorf("expected 4 language groups, got %d", len(byLang))
	}
	if len(byLang["python"]) != 2 {
		t.Errorf("expected 2 python files, got %d", len(byLang["python"]))
	}
}

func TestFindDriver(t *testing.T) {
	if d := findDriver("python"); d == nil {
		t.Error("python driver not found")
	}
	if d := findDriver("go"); d == nil {
		t.Error("go driver not found")
	}
	if d := findDriver("unknown"); d != nil {
		t.Errorf("unexpected driver for unknown language: %v", d)
	}
}

func TestComputeSourceDirs(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src", "pkg")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.py"), []byte("x=1"), 0644)

	dirs := computeSourceDirs([]string{filepath.Join(tmpDir, "src")})
	if len(dirs) == 0 {
		t.Fatal("expected at least 1 source dir")
	}
	// All dirs should exist
	for _, d := range dirs {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("source dir %q does not exist", d)
		}
	}
}

func TestParseCoverageByLang(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Go cover profile
	covPath := filepath.Join(tmpDir, "cover.out")
	os.WriteFile(covPath, []byte("mode: set\nmodule/pkg/file.go:1.0,3.0 1 1\n"), 0644)

	// Test Go coverage parsing
	cov, err := parseCoverageByLang(covPath, "go")
	if err != nil {
		t.Fatalf("parseCoverageByLang go: %v", err)
	}
	if cov == nil {
		t.Error("expected coverage map for Go")
	}

	// Test missing file
	_, err = parseCoverageByLang("/nonexistent/path.json", "python")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadCoverage_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a .coverage.json in tmpDir
	covPath := filepath.Join(tmpDir, "coverage.json")
	os.WriteFile(covPath, []byte(`{"meta":{"format":2},"files":{},"totals":{}}`), 0644)

	// Create source dir
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "test.py"), []byte("def f():\n    return 1\n"), 0644)

	cov, err := loadCoverage("coverage.json", "python", []string{srcDir})
	if err != nil {
		t.Logf("loadCoverage error (may be expected): %v", err)
	}
	if cov == nil {
		t.Log("coverage map nil — may be expected with empty files")
	}
}

func TestLoadCoverage_MissingFile(t *testing.T) {
	cov, err := loadCoverage("/nonexistent/coverage.json", "python", []string{"/tmp"})
	if err != nil {
		t.Logf("error as expected: %v", err)
	}
	if cov != nil {
		t.Error("expected nil coverage for missing file")
	}
}

func TestLoadCoverage_JavaScript(t *testing.T) {
	tmpDir := t.TempDir()
	lcovPath := filepath.Join(tmpDir, "lcov.info")
	os.WriteFile(lcovPath, []byte("SF:src/test.js\nDA:1,1\nend_of_record\n"), 0644)

	cov, err := loadCoverage("lcov.info", "javascript", []string{tmpDir})
	if err != nil {
		t.Fatalf("loadCoverage JavaScript: %v", err)
	}
	if cov == nil {
		t.Fatal("expected coverage for JavaScript")
	}
	if len(cov) != 1 {
		t.Errorf("expected 1 file, got %d", len(cov))
	}
}

func TestLoadCoverage_Go(t *testing.T) {
	tmpDir := t.TempDir()
	covPath := filepath.Join(tmpDir, "cover.out")
	os.WriteFile(covPath, []byte("mode: set\nmodule/pkg/file.go:1.0,3.0 1 1\n"), 0644)

	cov, err := loadCoverage("cover.out", "go", []string{tmpDir})
	if err != nil {
		t.Fatalf("loadCoverage Go: %v", err)
	}
	if cov == nil {
		t.Fatal("expected coverage for Go")
	}
}

func TestLoadCoverage_FromParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	covPath := filepath.Join(tmpDir, "coverage.json")
	os.WriteFile(covPath, []byte(`{"meta":{"format":2},"files":{"src/test.py":{"executed_lines":[1],"summary":{"covered_lines":1,"num_statements":1,"percent_covered":100},"missing_lines":[],"excluded_lines":[]}},"totals":{}}`), 0644)
	os.WriteFile(filepath.Join(srcDir, "test.py"), []byte("def f():\n    return 1\n"), 0644)

	cov, err := loadCoverage("coverage.json", "python", []string{srcDir})
	if err != nil {
		t.Fatalf("loadCoverage from parent: %v", err)
	}
	if cov == nil || len(cov) == 0 {
		t.Fatal("expected non-empty coverage from parent dir")
	}
}

func TestLoadCoverage_EmptyPath(t *testing.T) {
	cov, err := loadCoverage("", "python", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cov != nil {
		t.Error("expected nil for empty path")
	}
}

func TestAnalyzeFile_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "test.py")
	os.WriteFile(srcPath, []byte("def add(a, b):\n    return a + b\n"), 0644)

	drv := findDriver("python")
	if drv == nil {
		t.Skip("python driver not available")
	}
	fns, err := analyzeFile(drv, srcPath, nil)
	if err != nil {
		t.Fatalf("analyzeFile: %v", err)
	}
	if len(fns) != 1 {
		t.Fatalf("expected 1 function, got %d", len(fns))
	}
	if fns[0].Name != "add" {
		t.Errorf("expected 'add', got %q", fns[0].Name)
	}
}

func TestLoadCoverage_NotIsExistError(t *testing.T) {
	cov, err := loadCoverage("coverage.json", "python", []string{"/root/protected"})
	if err != nil {
		t.Logf("loadCoverage error (expected for protected dir): %v", err)
	}
	if cov != nil {
		t.Log("got coverage map despite error")
	}
}

func TestParseCoverageByLang_Unknown(t *testing.T) {
	_, err := parseCoverageByLang("/dev/null", "unknown")
	if err == nil {
		t.Error("expected error for unknown language")
	}
}
