package utils

import "strings"

func ArrSubStr(arr []string, sub string) bool {
	for _, item := range arr {
		if strings.Contains(item, sub) {
			return true
		}
	}
	return false
}
