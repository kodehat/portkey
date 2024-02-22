package utils

import "fmt"

func PageTitle(pageName string, title string) string {
	return fmt.Sprintf("%s - %s", pageName, title)
}
