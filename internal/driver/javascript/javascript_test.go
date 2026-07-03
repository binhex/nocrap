package javascript_test

import (
	"os"
	"testing"

	"nocrap/internal/driver"
	"nocrap/internal/driver/javascript"
)

func TestFindFunctions(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/javascript/simple.js")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := javascript.New()
	funcs, err := d.FindFunctions(source, "testdata/javascript/simple.js")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	if len(funcs) < 7 {
		t.Errorf("expected at least 7 functions, got %d", len(funcs))
		for _, f := range funcs {
			t.Logf("  %s @ line %d", f.Name, f.StartLine)
		}
	}

	findFunc := func(name string) *driver.Function {
		for i := range funcs {
			if funcs[i].Name == name {
				return &funcs[i]
			}
		}
		return nil
	}

	if f := findFunc("add"); f == nil {
		t.Error("add function not found")
	}
	if f := findFunc("fetchData"); f == nil {
		t.Error("fetchData async function not found")
	}
	if f := findFunc("Calculator.add"); f == nil {
		t.Error("Calculator.add method not found")
	}
}

func TestCalcComplexity_Branches(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/javascript/branches.js")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := javascript.New()
	funcs, err := d.FindFunctions(source, "testdata/javascript/branches.js")
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
	t.Logf("allBranches CC = %d", cc)
}
