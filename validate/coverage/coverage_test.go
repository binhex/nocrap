package coverage_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nocrap/internal/config"
	"nocrap/internal/engine"
	"nocrap/validate"
)

type expectedEntry struct {
	Process float64 `json:"process"`
	Sum     float64 `json:"sum"`
}

type expectedData struct {
	LcovFull     expectedEntry `json:"lcov_full"`
	LcovHalf     expectedEntry `json:"lcov_half"`
	LcovNone     expectedEntry `json:"lcov_none"`
	CoverFull    expectedEntry `json:"cover_full"`
	CoverHalf    expectedEntry `json:"cover_half"`
	CoverPartial expectedEntry `json:"cover_partial"`
}

func loadExpected(t *testing.T) expectedData {
	t.Helper()
	data, err := os.ReadFile("expected.json")
	if err != nil {
		t.Fatalf("reading expected.json: %v", err)
	}
	var exp expectedData
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatalf("parsing expected.json: %v", err)
	}
	return exp
}

func TestLCOVCoverage(t *testing.T) {
	exp := loadExpected(t)

	tests := []struct {
		name    string
		variant string
		want    float64
	}{
		{"full", "full", exp.LcovFull.Process},
		{"half", "half", exp.LcovHalf.Process},
		{"none", "none", exp.LcovNone.Process},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Coverage.JavaScript = filepath.Join("fixtures", "lcov_"+tt.variant+".info")

			scores, err := engine.Analyze([]string{filepath.Join("fixtures", "source.js")}, cfg)
			if err != nil {
				t.Fatalf("Analyze: %v", err)
			}

			var covPct float64
			for _, s := range scores {
				if s.Name == "process" {
					covPct = s.CoveragePercent
					break
				}
			}
			if !validate.WithinTolerance(covPct, tt.want, 0.5) {
				t.Errorf("CoveragePercent for lcov_%s: got %.2f, want %.2f", tt.variant, covPct, tt.want)
			}
		})
	}
}

func loadExpectedJSON(t *testing.T) map[string]map[string]float64 {
	t.Helper()
	data, err := os.ReadFile("expected.json")
	if err != nil {
		t.Fatalf("reading expected.json: %v", err)
	}
	var expected map[string]map[string]float64
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatalf("parsing expected.json: %v", err)
	}
	return expected
}

func TestGcovCoverage(t *testing.T) {
	expected := loadExpectedJSON(t)

	variants := []string{"full", "half", "none"}
	for _, variant := range variants {
		t.Run(variant, func(t *testing.T) {
			covFile := fmt.Sprintf("fixtures/%s.gcov", variant)

			cfg := config.DefaultConfig()
			cfg.Coverage.C = covFile

			scores, err := engine.Analyze([]string{"fixtures_c/add.c"}, cfg)
			if err != nil {
				t.Fatalf("Analyze: %v", err)
			}

			for _, s := range scores {
				wantKey := "gcov_" + variant
				wantPct := expected[wantKey][s.Name]
				if !validate.WithinTolerance(s.CoveragePercent, wantPct, 0.5) {
					t.Errorf("%s coverage = %.1f%%, want %.1f%% (\u00b10.5)",
						s.Name, s.CoveragePercent, wantPct)
				}
				t.Logf("  %s: CC=%d, Cov=%.1f%%, CRAP=%.2f", s.Name, s.CC, s.CoveragePercent, s.CRAP)
			}
		})
	}
}

func TestGoCoverage(t *testing.T) {
	exp := loadExpected(t)

	tests := []struct {
		name    string
		variant string
		want    float64
	}{
		{"full", "full", exp.CoverFull.Sum},
		{"half", "half", exp.CoverHalf.Sum},
		{"partial", "partial", exp.CoverPartial.Sum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Coverage.Go = filepath.Join("fixtures", "cover_"+tt.variant+".out")

			scores, err := engine.Analyze([]string{filepath.Join("fixtures", "source.go")}, cfg)
			if err != nil {
				t.Fatalf("Analyze: %v", err)
			}

			// Debug: log all functions found
			for _, s := range scores {
				t.Logf("DEBUG func=%q file=%q cc=%d cov=%.1f%% lines=%d-%d",
					s.Name, s.File, s.CC, s.CoveragePercent, s.StartLine, s.EndLine)
			}

			var covPct float64
			for _, s := range scores {
				if s.Name == "fixtures.sum" || strings.HasSuffix(s.Name, ".sum") {
					covPct = s.CoveragePercent
					t.Logf("DEBUG matched func=%q coverage=%.1f%%", s.Name, covPct)
					break
				}
			}
			t.Logf("DEBUG final covPct=%.1f%% want=%.1f%%", covPct, tt.want)
			if !validate.WithinTolerance(covPct, tt.want, 0.5) {
				t.Errorf("CoveragePercent for cover_%s: got %.2f, want %.2f", tt.variant, covPct, tt.want)
			}
		})
	}
}
