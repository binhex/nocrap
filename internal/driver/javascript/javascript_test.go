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

func TestCalcComplexity_OptionalChain(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/javascript/optional_chain.js")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := javascript.New()
	funcs, err := d.FindFunctions(source, "testdata/javascript/optional_chain.js")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "optionalChaining" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("optionalChaining not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	t.Logf("optionalChaining CC = %d", cc)
	// Should count ?? and outermost ?. but not nested ?.
	if cc < 2 {
		t.Errorf("CC = %d, expected at least 2", cc)
	}
}

func TestCalcComplexity_OptionalChain_Deep(t *testing.T) {
	source := []byte("function deepChain(a, b, c) {\n    return a?.b?.c?.d ?? b?.x ?? c?.y;\n}\n")
	d := javascript.New()
	funcs, err := d.FindFunctions(source, "test.js")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "deepChain" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("deepChain not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	t.Logf("deepChain CC = %d", cc)
	// Should count 3 ?? operators (+3) and at most 1 outermost ?. per chain
	if cc < 3 {
		t.Errorf("CC = %d, expected at least 3", cc)
	}
}

func TestCalcComplexity_OptionalChain_Simple(t *testing.T) {
	source := []byte("function simple(a) {\n    return a?.b;\n}\n")
	d := javascript.New()
	funcs, err := d.FindFunctions(source, "test.js")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "simple" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("simple not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	t.Logf("simple CC = %d", cc)
	if cc < 1 {
		t.Errorf("CC = %d, expected at least 1", cc)
	}
}

func TestCalcComplexity_OptionalChain_Nested(t *testing.T) {
	// a?.b?.c — inner ?. should be deduped, only outermost counts
	source := []byte("function nested(a) {\n    return a?.b?.c;\n}\n")
	d := javascript.New()
	funcs, err := d.FindFunctions(source, "test.js")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "nested" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("nested not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	t.Logf("nested CC = %d", cc)
	// Only 1 (outermost ?.) should count
	if cc < 1 {
		t.Errorf("CC = %d, expected at least 1", cc)
	}
	if cc > 2 {
		t.Errorf("CC = %d, expected at most 2 for single chain", cc)
	}
}

func TestCalcComplexity_OptionalChain_Standalone(t *testing.T) {
	// Just a?.b on its own — no nesting, should count
	source := []byte("function standalone(a) {\n    const x = a?.b;\n}\n")
	d := javascript.New()
	funcs, err := d.FindFunctions(source, "test.js")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "standalone" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("standalone not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	t.Logf("standalone CC = %d", cc)
	if cc < 1 {
		t.Errorf("CC = %d, expected at least 1", cc)
	}
}
