package utils

import "strings"

// ArrSubStr checks if the given string sub is contained
// in one of the strings of the array arr.
func ArrSubStr(arr []string, sub string) bool {
	for _, item := range arr {
		if strings.Contains(item, sub) {
			return true
		}
	}
	return false
}
