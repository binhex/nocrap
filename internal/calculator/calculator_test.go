package calculator_test

import (
	"math"
	"testing"

	"nocrap/internal/calculator"
)

func TestCRAP(t *testing.T) {
	tests := []struct {
		name     string
		cc       int
		covPct   float64
		wantCRAP float64
	}{
		{"cc1_cov0", 1, 0.0, 2.0},
		{"cc1_cov100", 1, 100.0, 1.0},
		{"cc1_cov50", 1, 50.0, 1.125},
		{"cc5_cov0", 5, 0.0, 30.0},
		{"cc5_cov100", 5, 100.0, 5.0},
		{"cc5_cov80", 5, 80.0, 5.2},
		{"cc20_cov0", 20, 0.0, 420.0},
		{"cc0_cov0", 0, 0.0, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculator.CRAP(tt.cc, tt.covPct)
			if math.Abs(got-tt.wantCRAP) > 0.001 {
				t.Errorf("CRAP(%d, %.1f) = %.3f, want %.3f", tt.cc, tt.covPct, got, tt.wantCRAP)
			}
		})
	}
}
