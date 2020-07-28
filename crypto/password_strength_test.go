package crypto

import (
	"fmt"
	"testing"

	"github.com/foomo/shop/utils"
)

func TestCryptoPasswordStrength(t *testing.T) {
	passwords := []string{"mypassword", "summertablecactus+", "osome+#,,brassford"}
	for _, password := range passwords {
		fmt.Println("---------- Password:", password, "-------------")
		utils.PrintJSON(determinePasswordStrength(password, nil))
	}

	fmt.Println("---------- Password with userInput-------------")
	// Test with user input, expected score 1
	determinePasswordStrength(passwords[1], []string{"Table"})
	utils.PrintJSON(determinePasswordStrength(passwords[1], []string{"Table"}))

	//for i := 0; i < 1000; i++ {
	for _, password := range passwords {
		fmt.Println("\n---------- GetScore:", password, " -------------")
		score, err := GetPasswordScore(password, nil)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Password:", password, "\tScore:", score)
	}
	//}
	SetMinLength(11)
	SetMaxLength(15)
	fmt.Println("\n\n -- Added restriction min=11, max=15 --")
	for _, password := range passwords {
		fmt.Println("\n---------- GetScore:", password, " -------------")
		score, err := GetPasswordScore(password, nil)

		if err == nil {
			t.Fatal(err)
		}
		fmt.Println("Password:", password, "\tScore:", score, "Error:", err.Error())
	}
}
