package engine_test

import (
	"path/filepath"
	"strings"
	"testing"

	"nocrap/internal/config"
	"nocrap/internal/engine"
)

func TestAnalyze_Python(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Coverage.Python = filepath.Join("..", "..", "testdata", "python", ".coverage.json")

	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "python")}, cfg)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(scores) == 0 {
		t.Error("expected at least 1 function score")
	}
	for _, s := range scores {
		t.Logf("%s: CC=%d, Cov=%.1f%%, CRAP=%.2f", s.Name, s.CC, s.CoveragePercent, s.CRAP)
	}
}

func TestAnalyze_Python_NoCoverage(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Coverage.Python = "/nonexistent/path.json"

	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "python")}, cfg)
	if err != nil {
		t.Fatalf("Analyze no coverage: %v", err)
	}
	if len(scores) == 0 {
		t.Error("expected functions even without coverage")
	}
}

func TestAnalyze_Go(t *testing.T) {
	cfg := config.DefaultConfig()
	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "go")}, cfg)
	if err != nil {
		t.Fatalf("Analyze Go: %v", err)
	}
	if len(scores) == 0 {
		t.Error("expected at least 1 Go function")
	}
	t.Logf("Found %d Go functions", len(scores))
}

func TestAnalyze_JS(t *testing.T) {
	cfg := config.DefaultConfig()
	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "javascript")}, cfg)
	if err != nil {
		t.Fatalf("Analyze JS: %v", err)
	}
	if len(scores) == 0 {
		t.Error("expected at least 1 JS function")
	}
	t.Logf("Found %d JS functions", len(scores))
}

func TestAnalyze_TS(t *testing.T) {
	cfg := config.DefaultConfig()
	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "typescript")}, cfg)
	if err != nil {
		t.Fatalf("Analyze TS: %v", err)
	}
	if len(scores) == 0 {
		t.Error("expected at least 1 TS function")
	}
	t.Logf("Found %d TS functions", len(scores))
}

func TestAnalyze_WithLang(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Lang = "python"
	cfg.Coverage.Python = filepath.Join("..", "..", "testdata", "python", ".coverage.json")

	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "python")}, cfg)
	if err != nil {
		t.Fatalf("Analyze with --lang: %v", err)
	}
	countPy := 0
	for _, s := range scores {
		if len(s.File) > 3 && s.File[len(s.File)-3:] == ".py" {
			countPy++
		}
	}
	if countPy == 0 {
		t.Error("expected at least 1 Python file")
	}
}

func TestAnalyze_WithLangGo(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Lang = "go"

	scores, err := engine.Analyze([]string{filepath.Join("..", "..", "testdata", "go")}, cfg)
	if err != nil {
		t.Fatalf("Analyze with --lang go: %v", err)
	}
	if len(scores) == 0 {
		t.Error("expected at least 1 Go function")
	}
	for _, s := range scores {
		if !strings.HasSuffix(s.File, ".go") {
			t.Errorf("expected only .go files, got %s", s.File)
		}
	}
}

func TestAnalyze_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	scores, err := engine.Analyze([]string{tmpDir}, cfg)
	if err != nil {
		t.Fatalf("Analyze empty dir: %v", err)
	}
	if len(scores) != 0 {
		t.Errorf("expected 0 functions from empty dir, got %d", len(scores))
	}
}

func TestAnalyze_MultiLanguage(t *testing.T) {
	cfg := config.DefaultConfig()
	scores, err := engine.Analyze([]string{
		filepath.Join("..", "..", "testdata", "python"),
		filepath.Join("..", "..", "testdata", "go"),
	}, cfg)
	if err != nil {
		t.Fatalf("Analyze multi: %v", err)
	}
	if len(scores) == 0 {
		t.Error("expected functions from multiple languages")
	}
	// Should have both .py and .go files
	hasPy := false
	hasGo := false
	for _, s := range scores {
		if strings.HasSuffix(s.File, ".py") {
			hasPy = true
		}
		if strings.HasSuffix(s.File, ".go") {
			hasGo = true
		}
	}
	if !hasPy || !hasGo {
		t.Errorf("expected .py and .go files, got hasPy=%v hasGo=%v", hasPy, hasGo)
	}
}
