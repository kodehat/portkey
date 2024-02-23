package utils

import "fmt"

func BuildAddr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
