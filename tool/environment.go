package environment

import (
	"syscall"
)

func GetOr(string2 string, string3 string) string {
	if v, ok := syscall.Getenv(string2); ok {
		return v
	}
	return string3
}
