package go_driver_test

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/driver"
	goDriver "nocrap/internal/driver/go"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "..", "testdata", "go", name)
}

func TestFindFunctions(t *testing.T) {
	source, err := os.ReadFile(testdataPath("simple.go"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := goDriver.New()
	funcs, err := d.FindFunctions(source, "testdata/go/simple.go")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	if len(funcs) < 4 {
		t.Errorf("expected at least 4 functions, got %d", len(funcs))
		for _, f := range funcs {
			t.Logf("  %s @ line %d", f.Name, f.StartLine)
		}
	}

	foundGreet := false
	for _, f := range funcs {
		if f.Name == "Greeter.Greet" {
			foundGreet = true
			break
		}
	}
	if !foundGreet {
		t.Error("Greeter.Greet method not found")
	}
}

func TestCalcComplexity_AllBranches(t *testing.T) {
	source, err := os.ReadFile(testdataPath("branches.go"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	d := goDriver.New()
	funcs, err := d.FindFunctions(source, "testdata/go/branches.go")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	var fn *driver.Function
	for i := range funcs {
		if funcs[i].Name == "branches.AllBranches" {
			fn = &funcs[i]
			break
		}
	}
	if fn == nil {
		t.Fatal("AllBranches not found")
	}

	cc, err := d.CalcComplexity(source, *fn)
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}

	if cc < 8 {
		t.Errorf("CC = %d, expected at least 8", cc)
	}
	t.Logf("AllBranches CC = %d", cc)
}
