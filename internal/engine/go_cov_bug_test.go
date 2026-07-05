package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/config"
	"nocrap/internal/engine"
)

func TestGoCoverage_FunctionSignatureInflatesDenominator(t *testing.T) {
	// Bug: computeCoverage counts ALL physical lines in the range for the numerator
	// (checking CoveredLines[ln]), but countExecutableLines for the denominator
	// excludes blank lines, comment-only lines, and brace-only lines.
	//
	// Since cover.out blocks may include closing-brace lines (which don't
	// pass isCodeLine), but the function signature line is NEVER in a block,
	// the numerator and denominator are systematically misaligned.
	//
	// Worse: TotalLines includes lines that isCodeLine would exclude (like braces),
	// but covered also includes them because blocks covering e.g. lines 5-6
	// include line 6 (a closing brace) in CoveredLines. Meanwhile the function
	// signature line is in totalLines but NOT in covered.
	//
	// This test uses a function where the signature-to-content ratio is high
	// (many short functions) to make the gap visible.

	tmpDir := t.TempDir()

	// Short functions with dense signatures: function signature inflates
	// totalLines but is never covered by blocks
	srcFile := filepath.Join(tmpDir, "thin.go")
	srcContent := []byte(`package thin

func a(i int) int { 
	return i + 1 
}

func b(i int) int { 
	return i + 2 
}

func c(i int) int { 
	return i + 3 
}
`)
	if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	// All functions fully covered — all lines except function signatures
	covContent := `mode: set
thin.go:4.2,4.22 1 1
thin.go:8.2,8.22 1 1
thin.go:12.2,12.22 1 1
`
	covFile := filepath.Join(tmpDir, "cover.out")
	if err := os.WriteFile(covFile, []byte(covContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Coverage.Go = covFile

	scores, err := engine.Analyze([]string{srcFile}, cfg)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	for _, s := range scores {
		t.Logf("%s: CC=%d, Cov=%.1f%%, CRAP=%.2f, lines=%d-%d",
			s.Name, s.CC, s.CoveragePercent, s.CRAP, s.StartLine, s.EndLine)

		if s.CoveragePercent != 100.0 {
			// Function has 1 executable line (the return) and 1 signature line.
			// The return is covered; the signature is not.
			// totalLines = 2 (signature + return), covered = 1 (return)
			// coverage = 1/2 = 50% — WRONG! Should be 100% because ALL
			// statements are covered.
			t.Errorf("%s: coverage = %.1f%%, want 100.0%% — "+
				"function signature line pollutes denominator",
				s.Name, s.CoveragePercent)
		}
	}
}
