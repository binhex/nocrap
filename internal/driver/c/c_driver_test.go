package c_test

import (
	"testing"

	"nocrap/internal/driver/c"
)

func TestFindFunctions(t *testing.T) {
	source := []byte(`int add(int a, int b) {
    return a + b;
}

int max(int a, int b) {
    if (a > b) return a;
    if (a < b) return b;
    return 0;
}
`)
	d := c.New()
	funcs, err := d.FindFunctions(source, "test.c")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(funcs))
	}
	if funcs[0].Name != "add" {
		t.Errorf("first function name = %q, want %q", funcs[0].Name, "add")
	}
	if funcs[1].Name != "max" {
		t.Errorf("second function name = %q, want %q", funcs[1].Name, "max")
	}
}

func TestCalcComplexity_Branches(t *testing.T) {
	source := []byte(`int max(int a, int b) {
    if (a > b) return a;
    if (a < b) return b;
    return 0;
}
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	// Base(1) + if(2) = 3
	if cc != 3 {
		t.Errorf("CC = %d, want 3", cc)
	}
}

func TestCalcComplexity_Switch(t *testing.T) {
	source := []byte(`int classify(int x) {
    switch (x) {
        case 1: return 1;
        case 2: return 2;
        case 3: return 3;
        default: return 0;
    }
}
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	// Base(1) + case(4, including default) = 5
	if cc != 5 {
		t.Errorf("CC = %d, want 5", cc)
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
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	// Base(1) + while(1) + do(1) + ternary(1) = 4
	if cc != 4 {
		t.Errorf("CC = %d, want 4", cc)
	}
}

func TestCalcComplexity_BooleanOps(t *testing.T) {
	source := []byte(`int check(int a, int b, int c) {
    if (a && b || c) {
        return 1;
    }
    return 0;
}
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	cc, err := d.CalcComplexity(source, funcs[0])
	if err != nil {
		t.Fatalf("CalcComplexity: %v", err)
	}
	// Base(1) + if(1) + &&(1) + ||(1) = 4
	if cc != 4 {
		t.Errorf("CC = %d, want 5", cc)
	}
}

func TestFindFunctions_FunctionPtr(t *testing.T) {
	source := []byte(`void apply(int (*fn)(int)) {
    fn(42);
}
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	if len(funcs) != 1 || funcs[0].Name != "apply" {
		t.Fatalf("expected 'apply', got %d funcs", len(funcs))
	}
}

func TestCalcComplexity_ForAndConditional(t *testing.T) {
	source := []byte(`int sum(int* arr, int n) {
    int s = 0;
    for (int i = 0; i < n; i++) {
        s += arr[i];
    }
    return n > 0 ? s : 0;
}
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	cc, _ := d.CalcComplexity(source, funcs[0])
	// Base(1) + for(1) + ternary(1) = 3
	if cc != 3 {
		t.Errorf("CC = %d, want 3", cc)
	}
}

func TestFindFunctions_MultipleFuncs(t *testing.T) {
	source := []byte(`int a() { return 1; }
int b() { return 2; }
int c() { return 3; }
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	if len(funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(funcs))
	}
}

func TestFindFunctions_HeaderStyle(t *testing.T) {
	source := []byte(`static int hidden(int x) { return x * 2; }
int visible(int x) { return x; }
void* create(void) { return 0; }
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	if len(funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(funcs))
	}
}

func TestFindFunctions_ParamVariants(t *testing.T) {
	source := []byte(`void empty() {}
int one(int a) { return a; }
int two(int a, int b) { return a+b; }
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	if len(funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(funcs))
	}
}

func TestFindFunctions_DeclaratorEdgeCases(t *testing.T) {
	source := []byte(`const char* getName(void) { return "test"; }
unsigned long getValue(void) { return 42; }
static inline int fast(int x) { return x; }
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	if len(funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(funcs))
	}
}

func TestFindFunctions_RefAndVolatile(t *testing.T) {
	source := []byte(`int* getPtr(volatile int* p) { return (int*)p; }
double calc(register int n) { return n * 1.0; }
`)
	d := c.New()
	funcs, _ := d.FindFunctions(source, "test.c")
	if len(funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(funcs))
	}
}
