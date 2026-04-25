package util

import "strings"

// ParseVariable Format var="value"
func ParseVariable(data string) (string, string) {
	args := strings.Split(data, "=")
	if len(args) != 2 {
		return "", ""
	}

	return args[0], strings.Trim(args[1], "\"")
}
