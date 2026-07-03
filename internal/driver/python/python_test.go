package python_test

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/driver"
	"nocrap/internal/driver/python"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "..", "testdata", "python", name)
}

func TestFindFunctions(t *testing.T) {
	fp := testdataPath("simple.py")
	source, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := python.New()
	funcs, err := d.FindFunctions(source, fp)
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	if len(funcs) < 6 {
		t.Errorf("expected at least 6 functions, got %d: %v", len(funcs), names(funcs))
	}

	find := func(name string) *driver.Function {
		for i := range funcs {
			if funcs[i].Name == name {
				return &funcs[i]
			}
		}
		return nil
	}

	add := find("add")
	if add == nil {
		t.Fatal("add function not found")
	}
	if add.CoverageStartLine < 5 {
		t.Errorf("add.CoverageStartLine = %d, should exclude docstring/module docstring", add.CoverageStartLine)
	}

	calcInit := find("Calculator.__init__")
	if calcInit == nil {
		t.Fatal("Calculator.__init__ not found")
	}
	if calcInit.Package != "Calculator" {
		t.Errorf("Package = %q, want %q", calcInit.Package, "Calculator")
	}
}

func TestCalcComplexity_Branches(t *testing.T) {
	fp := testdataPath("branches.py")
	source, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := python.New()
	funcs, err := d.FindFunctions(source, fp)
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var allBranches *driver.Function
	for i := range funcs {
		if funcs[i].Name == "all_branches" {
			allBranches = &funcs[i]
			break
		}
	}
	if allBranches == nil {
		t.Fatal("all_branches function not found")
	}

	cc, err := d.CalcComplexity(source, *allBranches)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	// all_branches: if(1) + elif(1) + while(1) + for(1) + except(2) + with(1) + and(1) + or(1) + base(1)
	if cc < 8 {
		t.Errorf("CC for all_branches = %d, expected at least 8", cc)
	}
	t.Logf("all_branches CC = %d", cc)
}

func TestCalcComplexity_EmptyBody(t *testing.T) {
	source := []byte("def empty():\n    pass\n")
	d := python.New()
	funcs, err := d.FindFunctions(source, "test.py")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	if cc != 1 {
		t.Errorf("CC for empty function = %d, want 1", cc)
	}
}

func names(funcs []driver.Function) []string {
	n := make([]string, len(funcs))
	for i, f := range funcs {
		n[i] = f.Name
	}
	return n
}
