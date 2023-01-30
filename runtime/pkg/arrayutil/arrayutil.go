package arrayutil

func Dedupe[E comparable](array []E) []E {
	var deduped []E
	dedupeMap := make(map[E]bool)

	for _, e := range array {
		if _, ok := dedupeMap[e]; ok {
			continue
		}
		dedupeMap[e] = true
		deduped = append(deduped, e)
	}

	return deduped
}

func Contains[E comparable](array []E, c E) bool {
	for _, e := range array {
		if e == c {
			return true
		}
	}
	return false
}

func Delete[E comparable](array []E, c E) []E {
	for i, e := range array {
		if e == c {
			return append(array[0:i], array[i+1:]...)
		}
	}
	return array
}

// Reverse reverses a slice in-place
func Reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// RangeInt returns creates a slice with integers from start to end(excluding)
// the range can also be reversed using rev flag
func RangeInt(start, end int, rev bool) []int {
	if end <= start {
		return make([]int, 0)
	}

	result := make([]int, end-start)
	for i := start; i < end; i++ {
		result[i-start] = i
	}

	if rev {
		Reverse(result)
	}
	return result
}
