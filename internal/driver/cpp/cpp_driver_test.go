package cpp_test

import (
	"strings"
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

func TestFindFunctions_OutOfClass(t *testing.T) {
	source := []byte(`class Calculator {
public:
    int add(int a, int b);
};

int Calculator::add(int a, int b) {
    return a + b;
}
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "test.cpp")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("expected 1 function (out-of-class definition), got %d", len(funcs))
	}
	if funcs[0].Name == "" {
		t.Errorf("function name should not be empty for out-of-class definition")
	}
	t.Logf("out-of-class function name: %q", funcs[0].Name)
}

func TestFindFunctions_OperatorAndDestructor(t *testing.T) {
	source := []byte(`class Foo {
public:
    int operator+(int other) { return 1; }
    ~Foo() { }
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
	if funcs[0].Name != "Foo::operator+" {
		t.Errorf("operator name = %q, want %q", funcs[0].Name, "Foo::operator+")
	}
	if funcs[1].Name != "Foo::~Foo" {
		t.Errorf("destructor name = %q, want %q", funcs[1].Name, "Foo::~Foo")
	}
}

func TestCalcComplexity_WhileDoConditional(t *testing.T) {
	source := []byte(`int loop(int n) {
    int s = 0;
    while (n > 0) {
        s += n;
        n--;
    }
    do {
        s++;
    } while (s < 10);
    return n > 0 ? s : 0;
}
`)
	d := cpp.New()
	funcs, _ := d.FindFunctions(source, "test.cpp")
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	// Base(1) + while(1) + do(1) + ternary(1) = 4
	if cc != 4 {
		t.Errorf("CC = %d, want 4", cc)
	}
}

func TestFindFunctions_NestedClass(t *testing.T) {
	source := []byte(`class Outer {
public:
    class Inner {
    public:
        int getValue() { return 42; }
    };
};
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "test.cpp")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	if funcs[0].Name != "Outer::Inner::getValue" {
		t.Errorf("function name = %q, want %q", funcs[0].Name, "Outer::Inner::getValue")
	}
}

func TestFindFunctions_ConversionOperator(t *testing.T) {
	source := []byte(`class Bool {
public:
    operator bool() const { return true; }
};
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "test.cpp")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	if !strings.Contains(funcs[0].Name, "operator") {
		t.Errorf("function name should contain 'operator', got %q", funcs[0].Name)
	}
	t.Logf("conversion operator name: %q", funcs[0].Name)
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
