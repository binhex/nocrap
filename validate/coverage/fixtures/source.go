package fixtures

func sum(items []int) int {
	count := 0
	for _, x := range items {
		if x > 0 {
			count += x
		}
		if x < 0 {
			count += x
		}
	}
	return count
}
