package utils

import "testing"

// TestArrSubStr calls utils.ArrSubStr with some strings, checking
// for a valid return value.
func TestArrSubStr(t *testing.T) {
	strings := []string{"first", "second", "third"}
	actual := ArrSubStr(strings, "first")
	if !actual {
		t.Fatalf(`ArrSubStr(["first", "second", "third"], "first") == false`)
	}
}

// TestArrSubStrEmpty calls utils.ArrSubStr with an empty array,
// checking for no error.
func TestArrSubStrEmpty(t *testing.T) {
	strings := []string{}
	actual := ArrSubStr(strings, "first")
	if actual {
		t.Fatalf(`ArrSubStr([], "first") == true`)
	}
}
