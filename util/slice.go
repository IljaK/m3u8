package util

func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RemoveString(s []string, index int) {
	if index >= len(s) || index < 0 {
		return
	}
	last := len(s[index]) - 1
	s[index] = s[last]
	s = s[:last]
}
