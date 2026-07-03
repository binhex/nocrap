package fixtures

// NoBranches exercises straight-line code (CC=1).
func NoBranches(x int) int {
	return x + 1
}

// SingleIf exercises one if/else (CC=2).
func SingleIf(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}

// IfElseIf exercises if/else-if/else chain (CC=3).
func IfElseIf(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	} else {
		return 0
	}
}

// NestedIf exercises if inside if (CC=3).
func NestedIf(x, y int) int {
	if x > 0 {
		if y > 0 {
			return 1
		}
	}
	return 0
}

// ForLoop exercises a simple for loop (CC=2).
func ForLoop(lst []int) int {
	for i := 0; i < len(lst); i++ {
	}
	return 0
}

// ForWithIf exercises a for loop containing an if (CC=3).
func ForWithIf(lst []int) int {
	for i := 0; i < len(lst); i++ {
		if lst[i] > 0 {
			return lst[i]
		}
	}
	return 0
}

// WhileLoop exercises a for loop used as while (CC=2).
func WhileLoop(n int) int {
	for n > 0 {
		n--
	}
	return n
}

// TryCatch is a stub — Go has no try/catch; skipped via expected.json.
func TryCatch() int {
	return 0
}

// BooleanOps exercises boolean operators in a condition (CC=4).
func BooleanOps(a, b, c bool) int {
	if a && b || c {
		return 1
	}
	return 0
}

// EarlyReturn exercises multiple guard clauses (CC=3).
func EarlyReturn(val int) int {
	if val < 0 {
		return -1
	}
	if val == 0 {
		return 0
	}
	return 1
}

// Ternary uses if/else since Go has no ternary operator (CC=2).
func Ternary(x bool, a, b int) int {
	if x {
		return a
	}
	return b
}

// SwitchCase exercises a switch with 2 cases + default (CC=4 in Go).
func SwitchCase(x int) int {
	switch x {
	case 1:
		return 10
	case 2:
		return 20
	default:
		return 0
	}
}
