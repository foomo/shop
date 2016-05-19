package crypto

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type Crypto struct {
	HashedPassword []byte
	Salt           []byte
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func newSalt() ([]byte, error) {
	n := 10
	salt := make([]byte, n)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// HashPassword returns the hash for the password and the associated salt
func HashPassword(password string) (*Crypto, error) {
	salt, err := newSalt()
	if err != nil {
		return nil, err
	}
	passwordBytes := []byte(password)
	// Add salt to password
	passwordBytes = append(passwordBytes, salt...)
	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost+2)
	if err != nil {
		return nil, err
	}

	crypto := &Crypto{
		HashedPassword: hash,
		Salt:           salt,
	}
	return crypto, nil
}

func VerifyPassword(crypto *Crypto, password string) bool {
	passwordBytes := []byte(password)
	// Add salt to password
	passwordBytes = append(passwordBytes, crypto.Salt...)
	return bcrypt.CompareHashAndPassword(crypto.HashedPassword, passwordBytes) == nil
}
