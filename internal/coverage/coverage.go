// Package coverage parses language-specific coverage formats into a unified
// representation: a map from file path to the set of covered line numbers.
package coverage

// CoverageData holds the covered (executed) line numbers for a single source file.
// Line numbers are 1-based, matching standard coverage tool output.
type CoverageData struct {
	CoveredLines map[int]bool
	TotalLines   int // total executable lines discovered (may be inferred)
}

// CoverageMap maps normalized file paths to their coverage data.
type CoverageMap map[string]*CoverageData
