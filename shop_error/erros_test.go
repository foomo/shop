package shop_error

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	err := errors.New("not found")
	if !IsError(err, ErrorNotInDatabase) {
		t.Fail()
	}
	if IsError(err, "ErrorBla") {
		t.Fail()
	}
}
