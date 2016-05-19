package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Crypto struct {
	HashedPassword []byte
	Salt           []byte
}

func NewSalt() ([]byte, error) {
	n := 10
	salt := make([]byte, n)
	_, err := rand.Read(salt)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}

	return salt, nil
}

// HashPassword returns the hash for the password and the associated salt
func HashPassword(password string) (*Crypto, error) {
	salt, err := NewSalt()
	if err != nil {
		return nil, err
	}
	passwordBytes := []byte(password)
	// add salt to password
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
	// add salt to password
	passwordBytes = append(passwordBytes, crypto.Salt...)
	return bcrypt.CompareHashAndPassword(crypto.HashedPassword, passwordBytes) == nil
}
