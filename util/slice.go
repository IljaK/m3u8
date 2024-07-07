package util

func Contains[V comparable](s []V, e V) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RemoveAt[V any](s []V, index int, keepOrder bool) []V {
	if index >= len(s) || index < 0 {
		return s
	}
	if keepOrder {
		if len(s) > index+1 {
			return append(s[:index], s[index+1:]...)
		}
		return s[:index]
	}
	last := len(s) - 1
	s[index] = (s)[last]
	return s[:last]
}

func Remove[V comparable](s []V, item V, keepOrder bool) []V {
	for i := 0; i < len(s); i++ {
		if item == s[i] {
			return RemoveAt(s, i, keepOrder)
		}
	}
	return s
}

func RemoveAll[V comparable](s []V, item V, keepOrder bool) []V {
	for i := len(s) - 1; i >= 0; i-- {
		if item == s[i] {
			s = RemoveAt(s, i, keepOrder)
		}
	}
	return s
}

func RemoveDuplicates[V comparable](s []V, keepOrder bool) []V {
	for n := len(s) - 1; n >= 0; n-- {
		for i := n - 1; i >= 0; i-- {
			if s[i] == s[n] {
				s = RemoveAt(s, n, keepOrder)
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

func Clone[S ~[]E, E any](s S) S {
	// Preserve nil in case it matters.
	if s == nil {
		return nil
	}
	c := make([]E, len(s), len(s))
	for i := 0; i < len(s); i++ {
		c[i] = s[i]
	}
	return c
}

func HasMatchAny[V comparable](arr1 []V, arr2 []V) bool {
	for i := 0; i < len(arr1); i++ {
		for j := 0; j < len(arr2); j++ {
			if arr1[i] == arr2[j] {
				return true
			}
		}
	}
	return false
}
