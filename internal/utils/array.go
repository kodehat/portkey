package utils

import "strings"

// ArrSubStr checks if the given string sub is contained
// in one of the strings of the array arr. The comparison is case-insensitive.
func ArrSubStr(arr []string, sub string) bool {
	lowerSub := strings.ToLower(sub)
	for _, item := range arr {
		if strings.Contains(strings.ToLower(item), lowerSub) {
			return true
		}
	}
	return false
}
