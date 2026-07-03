package typescript_test

import (
	"os"
	"testing"

	"nocrap/internal/driver"
	"nocrap/internal/driver/typescript"
)

func TestFindFunctions_TS(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/typescript/simple.ts")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := typescript.New()
	funcs, err := d.FindFunctions(source, "testdata/typescript/simple.ts")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	if len(funcs) < 4 {
		t.Errorf("expected at least 4 functions, got %d", len(funcs))
		for _, f := range funcs {
			t.Logf("  %s @ line %d", f.Name, f.StartLine)
		}
	}
}

func TestCalcComplexity_Branches(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/typescript/branches.ts")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := typescript.New()
	funcs, err := d.FindFunctions(source, "testdata/typescript/branches.ts")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "allBranches" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("allBranches not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	if cc < 10 {
		t.Errorf("CC = %d, expected at least 10", cc)
	}
	t.Logf("allBranches TS CC = %d", cc)
}
