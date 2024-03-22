package utils

import "fmt"

func NewError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
