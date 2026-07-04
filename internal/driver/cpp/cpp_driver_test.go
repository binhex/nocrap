package cpp_test

import (
	"testing"

	"nocrap/internal/driver/cpp"
)

func TestFindFunctions(t *testing.T) {
	source := []byte(`int add(int a, int b) {
    return a + b;
}

class Calculator {
public:
    int add(int a, int b) {
        return a + b;
    }
};
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "test.cpp")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(funcs))
	}
	if funcs[0].Name != "add" {
		t.Errorf("first function = %q, want %q", funcs[0].Name, "add")
	}
	if funcs[1].Name != "Calculator::add" {
		t.Errorf("second function = %q, want %q", funcs[1].Name, "Calculator::add")
	}
}

func TestCalcComplexity_Catch(t *testing.T) {
	source := []byte(`int safeDiv(int a, int b) {
    try {
        return a / b;
    } catch (int e) {
        return 0;
    } catch (...) {
        return -1;
    }
}
`)
	d := cpp.New()
	funcs, _ := d.FindFunctions(source, "test.cpp")
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	// Base(1) + catch(2) = 3
	if cc != 3 {
		t.Errorf("CC = %d, want 3", cc)
	}
}
