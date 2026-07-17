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

func TestFindFunctions_OutOfClassConversion(t *testing.T) {
	source := []byte(`class Bool {
public:
    operator bool() const;
};

Bool::operator bool() const { return true; }
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "test.cpp")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("expected 1 function (out-of-class conversion op), got %d", len(funcs))
	}
	t.Logf("out-of-class conversion operator: %q", funcs[0].Name)
}

func TestFindFunctions_Template(t *testing.T) {
	source := []byte(`template<typename T>
T max(T a, T b) {
    return a > b ? a : b;
}
`)
	d := cpp.New()
	funcs, _ := d.FindFunctions(source, "test.cpp")
	if len(funcs) < 1 {
		t.Fatalf("expected at least 1 function, got %d", len(funcs))
	}
}

func TestFindFunctions_Namespace(t *testing.T) {
	source := []byte(`namespace ns {
    int getValue() { return 42; }
}
`)
	d := cpp.New()
	funcs, _ := d.FindFunctions(source, "test.cpp")
	if len(funcs) < 1 {
		t.Fatalf("expected at least 1 function, got %d", len(funcs))
	}
}

func TestFindFunctions_MoreEdgeCases(t *testing.T) {
	source := []byte(`const int* getPtr() { return 0; }
static void helper() { }
inline int fast(int x) { return x; }
`)
	d := cpp.New()
	funcs, _ := d.FindFunctions(source, "test.cpp")
	if len(funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(funcs))
	}
}

func TestFindFunctions_PartialParseRecovery(t *testing.T) {
	// A C++ file with some unparseable content (unexpanded macro call)
	// should still return valid function definitions from the partial
	// tree-sitter AST instead of an error.
	source := []byte(`int add() { return 1; }
#define FOO(x) x
FOO(int)
int max() { return 2; }
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "test.cpp")
	if err != nil {
		t.Fatalf("FindFunctions should not error on partially parseable C++: %v", err)
	}
	if len(funcs) != 2 {
		t.Fatalf("expected 2 functions from partial parse, got %d", len(funcs))
	}
	if funcs[0].Name != "add" {
		t.Errorf("first function = %q, want %q", funcs[0].Name, "add")
	}
	if funcs[1].Name != "max" {
		t.Errorf("second function = %q, want %q", funcs[1].Name, "max")
	}
}

func TestFindFunctions_DataOnlyHeader(t *testing.T) {
	// Data-only headers that can't be parsed as standalone C++
	// should be skipped silently, returning no error and no functions.
	source := []byte(`/* 0x00 */ BT_NONXML, BT_NONXML, BT_NONXML, BT_NONXML,
    /* 0x04 */ BT_NONXML, BT_NONXML, BT_NONXML, BT_NONXML,
    /* 0x08 */ BT_NONXML, BT_S, BT_LF, BT_NONXML,
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "asciitab.h")
	if err != nil {
		t.Fatalf("FindFunctions on data-only header: %v", err)
	}
	if len(funcs) != 0 {
		t.Errorf("expected 0 functions from data-only header, got %d", len(funcs))
	}
}

func TestFindFunctions_NormalCppFunctionsStillWork(t *testing.T) {
	source := []byte(`int add(int a, int b) {
    return a + b;
}

class Foo {
public:
    int getValue() { return 42; }
};
`)
	d := cpp.New()
	funcs, err := d.FindFunctions(source, "normal.cpp")
	if err != nil {
		t.Fatalf("FindFunctions on normal C++: %v", err)
	}
	if len(funcs) < 2 {
		t.Errorf("expected at least 2 functions, got %d", len(funcs))
	}
}

func TestCalcComplexity_NestedAndBoolean(t *testing.T) {
	source := []byte(`int complex(int a, int b, int c) {
    if (a > 0 && b > 0) {
        if (c > 0) return 1;
    }
    return a || b ? 2 : 0;
}
`)
	d := cpp.New()
	funcs, _ := d.FindFunctions(source, "test.cpp")
	cc, _ := d.CalcComplexity(source, funcs[0])
	if cc < 4 {
		t.Errorf("CC = %d, want at least 4", cc)
	}
}
