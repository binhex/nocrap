package engine_test

import (
	"path/filepath"
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
