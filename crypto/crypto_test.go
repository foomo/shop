package crypto

import (
	"fmt"
	"testing"
)

func TestCryptoCreateSalt(t *testing.T) {
	for i := 0; i < 10; i++ {
		salt, err := NewSalt()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Salt: ", salt)
	}
}

func TestCryptoCreateAndVerifyHash(t *testing.T) {
	password := "totallySafeP@??m0rd!11"
	crypto, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Hash:", string(crypto.HashedPassword))

	wrongPassword := "thisissototallywrong"
	if !VerifyPassword(crypto, password) {
		t.Fail()
		fmt.Println("Correct password did not match against hash.")
	}

	if VerifyPassword(crypto, wrongPassword) {
		t.Fail()
		fmt.Println("Wrong password matched against hash.")
	}
}
