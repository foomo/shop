package shop_error

import "strings"

const (
	ErrorNotFound     = "not found"
	ErrorNotAvailable = "Not available"
)

// ErrorIs returns true, if err is of kind e, else false
func IsError(err error, e string) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), e)
}
