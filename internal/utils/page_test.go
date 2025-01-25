package utils

import "testing"

// TestPageTitle calls utils.PageTime with a page name and a title,
// checking for a valid return value.
func TestPageTitle(t *testing.T) {
	pageName := "Home"
	title := "portkey"
	actual := PageTitle(pageName, title)
	if actual != "Home - portkey" {
		t.Fatalf(`PageTitle("Home", "portkey") == %s, want "Home - portkey"`, actual)
	}
}
