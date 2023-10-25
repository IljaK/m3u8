package util

func Contains[V comparable](s []V, e V) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RemoveAt[V any](s []V, index int) []V {
	if index >= len(s) || index < 0 {
		return s
	}
	last := len(s) - 1
	s[index] = (s)[last]
	return s[:last]
}

func Remove[V comparable](s []V, item V) []V {
	for i := 0; i < len(s); i++ {
		if item == s[i] {
			return RemoveAt(s, i)
		}
	}
	return s
}

func RemoveAll[V comparable](s []V, item V) []V {
	for i := len(s) - 1; i >= 0; i-- {
		if item == s[i] {
			s = RemoveAt(s, i)
		}
	}
	return s
}

func RemoveDuplicates[V comparable](s []V) []V {
	for n := len(s) - 1; n >= 0; n-- {
		for i := n - 1; i >= 0; i-- {
			if s[i] == s[n] {
				s = RemoveAt(s, n)
				break
			}
		}
	}
	return s
}

func GetMapValuesArray[V comparable, T comparable](collection map[V]T) []T {
	arr := make([]T, 0, len(collection))

	for _, v := range collection {
		arr = append(arr, v)
	}
	return arr
}

func ToInterfaceArray[T any](array []T) []interface{} {
	values := make([]interface{}, 0, len(array))
	for _, arg := range array {
		values = append(values, arg)
	}
	return values
}
