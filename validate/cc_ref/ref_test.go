package cc_ref_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nocrap/internal/driver"
	goDriver "nocrap/internal/driver/go"
	"nocrap/internal/driver/typescript"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("fixtures", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return data
}

func loadExpected(t *testing.T, name string) map[string]int {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("reading expected %s: %v", name, err)
	}
	var expected map[string]int
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatalf("unmarshalling expected %s: %v", name, err)
	}
	return expected
}

func findFunction(funcs []driver.Function, name string) *driver.Function {
	for _, f := range funcs {
		// Handle both "name" and "package.name" formats
		if f.Name == name || strings.HasSuffix(f.Name, "."+name) {
			return &f
		}
	}
	return nil
}

func TestRefCCGo(t *testing.T) {
	source := loadFixture(t, "ref_go.go")
	expected := loadExpected(t, "expected_go.json")

	d := goDriver.New()
	funcs, err := d.FindFunctions(source, "fixtures/ref_go.go")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	for funcName, wantCC := range expected {
		fn := findFunction(funcs, funcName)
		if fn == nil {
			t.Fatalf("function %q not found in fixture", funcName)
		}
		gotCC, err := d.CalcComplexity(source, *fn)
		if err != nil {
			t.Fatalf("CalcComplexity(%q): %v", funcName, err)
		}
		if gotCC != wantCC {
			t.Errorf("CC for %q: got %d, want %d", funcName, gotCC, wantCC)
		}
	}
}

func TestRefCCTS(t *testing.T) {
	source := loadFixture(t, "ref_js.ts")
	expected := loadExpected(t, "expected_js.json")

	d := typescript.New()
	funcs, err := d.FindFunctions(source, "fixtures/ref_js.ts")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}

	for funcName, wantCC := range expected {
		fn := findFunction(funcs, funcName)
		if fn == nil {
			t.Fatalf("function %q not found in fixture", funcName)
		}
		gotCC, err := d.CalcComplexity(source, *fn)
		if err != nil {
			t.Fatalf("CalcComplexity(%q): %v", funcName, err)
		}
		if gotCC != wantCC {
			t.Errorf("CC for %q: got %d, want %d", funcName, gotCC, wantCC)
		}
	}
}
