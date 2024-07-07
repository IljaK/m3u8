package util

import (
	"strings"
)

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
		for n := len(result) - 1; n >= 0; n-- {
			sub := strings.Split(result[n], separators[i])
			if len(sub) > 1 {
				// Clone necessary here! without it "remain" changes on next line!
				remain := Clone(result[n+1:])

				sub = append(result[0:n], sub...)
				if len(result) > n+1 {
					sub = append(sub, remain...)
				}
				result = sub
			}
		}
	}
	return result
}

func TrimEmpty(arr []string, keepOrder bool) []string {

	for i := len(arr) - 1; i >= 0; i-- {
		arr[i] = strings.TrimSpace(arr[i])
		if arr[i] == "" {
			arr = RemoveAt(arr, i, keepOrder)
		}
	}
	return arr
}

func AddIfNotExist(arr []string, value string) []string {
	for i := 0; i < len(arr); i++ {
		if arr[i] == value {
			return arr
		}
	}
	return append(arr, value)
}
