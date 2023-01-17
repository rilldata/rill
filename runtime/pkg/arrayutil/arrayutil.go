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
