package branches

// AllBranches exercises all Go branching constructs.
func AllBranches(x, y int, items []int, ch chan int) int {
	result := 0

	// if/else
	if x > 0 {
		result = 1
	} else if x == 0 {
		result = 0
	} else {
		result = -1
	}

	// for (count)
	for i := 0; i < 10; i++ {
		result++
	}

	// for range
	for _, item := range items {
		result += item
	}

	// for (while style)
	for y > 0 {
		y--
	}

	// switch/case
	switch x {
	case 1:
		result = 10
	case 2:
		result = 20
	default:
		result = 0
	}

	// select/case
	select {
	case <-ch:
		result = 100
	default:
		result = 0
	}

	// && and || in conditions
	if x > 0 && y > 0 {
		result = 2
	}

	if x > 0 || y > 0 {
		result = 3
	}

	return result
}

// TypeSwitch exercises a type switch.
func TypeSwitch(v interface{}) string {
	switch t := v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	default:
		return "unknown"
	}
}
