// Package validate provides shared test helpers for the nocrap validation suite.
package validate

import "math"

// WithinTolerance returns true if actual is within tolerance of expected.
func WithinTolerance(actual, expected, tolerance float64) bool {
	return math.Abs(actual-expected) <= tolerance
}
