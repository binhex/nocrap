// Package calculator provides the language-agnostic CRAP score formula.
// CRAP = CC^2 * (1 - coverage_percent / 100)^3 + CC
package calculator

import "math"

// CRAP computes the Change Risk Anti-Patterns score.
// cc is cyclomatic complexity (must be >= 1).
// coveragePercent is line coverage percentage (0.0-100.0).
func CRAP(cc int, coveragePercent float64) float64 {
	if cc < 1 {
		cc = 1
	}
	covFactor := math.Pow(1.0-coveragePercent/100.0, 3)
	return float64(cc*cc)*covFactor + float64(cc)
}
