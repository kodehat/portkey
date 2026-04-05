package utils

import "fmt"

func PageTitle(pageName, title string) string {
	return fmt.Sprintf("%s - %s", pageName, title)
}
