package shop_error

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	err := errors.New("not found")
	if !ErrorIs(err, ErrorNotFound) {
		t.Fail()
	}
	if ErrorIs(err, "ErrorBla") {
		t.Fail()
	}
}
