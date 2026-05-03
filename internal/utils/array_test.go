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

func TestArrSubStr_CaseInsensitive(t *testing.T) {
	strings := []string{"GitHub", "GitLab"}
	if !ArrSubStr(strings, "hub") {
		t.Fatal("expected case-insensitive match")
	}
	if !ArrSubStr(strings, "HUB") {
		t.Fatal("expected case-insensitive match with uppercase query")
	}
	if ArrSubStr(strings, "bitbucket") {
		t.Fatal("expected no match for unrelated string")
	}
}
