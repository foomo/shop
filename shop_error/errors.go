package shop_error

import "strings"

const (
	ErrorAlreadyExists        = "already existing in db" // do not change, this string is returned by MongoDB
	ErrorNotInDatabase        = "not found"              // do not change, this string is returned by MongoDB
	ErrorNotFound             = "Error: Not Found: "
	ErrorRequiredFieldMissing = "Error: A required field is missing!"
)

// ErrorIs returns true, if err is of kind e, else false
func IsError(err error, e string) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), e)
}
