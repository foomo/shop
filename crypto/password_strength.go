package crypto

// This uses a Go-Port of "zxcvbn: realistic password strength estimation"
// See https://blogs.dropbox.com/tech/2012/04/zxcvbn-realistic-password-strength-estimation/ for further information

import (
	"errors"

	"git.bestbytes.net/Project-Globus-Services/utils"

	zxcbvn "github.com/nbutton23/zxcvbn-go"
	"github.com/nbutton23/zxcvbn-go/scoring"
)

var minLength int = -1
var maxLength int = -1

func SetMinLength(min int) {
	minLength = min
}
func SetMaxLength(max int) {
	maxLength = max
}

// DeterminePasswordStrength returns a detailed info about the strength of the given password
// @userInput e.g. user name. Given strings are matched against password to prohibit similarities between username and password
func DeterminePasswordStrength(password string, userInput []string) scoring.MinEntropyMatch {
	return zxcbvn.PasswordStrength(password, userInput)
}

// GetPasswordScore returns a score of 0 (poor), 1, 2, 3 or 4 (excellent) for the strength of the password
func GetPasswordScore(password string, userInput []string) (int, error) {
	if minLength != -1 && len(password) < minLength {
		return 0, errors.New("Password must have at least " + utils.IntToString(minLength) + " characters!")
	}
	if maxLength != -1 && len(password) > maxLength {
		return 0, errors.New("Password must be not longer than " + utils.IntToString(maxLength) + " characters!")
	}
	match := DeterminePasswordStrength(password, userInput)
	return match.Score, nil
}
