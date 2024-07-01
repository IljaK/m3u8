package util

import "strings"

func Concat(str1 string, str2 string, sep string) string {

	if str1 == "" {
		return str2
	}
	if str2 != "" {
		str1 += sep + str2
	}
	return str1
}

func JoinArrToString(sep string, target string, elems []string) string {
	for i := 0; i < len(elems); i++ {
		target = Concat(target, elems[i], sep)
	}
	return target
}

func JoinArr(sep string, elems []string) string {
	var result string
	for i := 0; i < len(elems); i++ {
		result = Concat(result, elems[i], sep)
	}
	return result
}

func Join(sep string, elems ...string) string {
	var result string
	for i := 0; i < len(elems); i++ {
		result = Concat(result, elems[i], sep)
	}
	return result
}

func SplitMultiple(target string, separators ...string) []string {
	result := []string{target}

	for i := 0; i < len(separators); i++ {
		var sub []string
		for n := 0; n < len(result); n++ {
			sub = append(strings.Split(result[n], separators[i]))
		}
		result = sub
	}

	out := make([]string, 0, 10)
	for i := 0; i < len(result); i++ {
		trimmed := strings.TrimSpace(result[i])
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}

func AddIfNotExist(arr []string, value string) []string {
	for i := 0; i < len(arr); i++ {
		if arr[i] == value {
			return arr
		}
	}
	return append(arr, value)
}

func HasNotEmpty(arr []string) bool {
	for i := 0; i < len(arr); i++ {
		if len(arr[i]) > 0 {
			return true
		}
	}
	return false
}
