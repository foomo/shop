package customer

import (
	"testing"

	"github.com/foomo/shop/test_utils"
)

func TestCredentials(t *testing.T) {

	test_utils.DropAllCollections()

	// Create credentials for a user
	email := "foo@bar.com"
	password := "123456"
	err := CreateCustomerCredentials(email, password)
	if err != nil {
		t.Fatal(err)
	}
	// Create credentials for another user
	email = "alice@bar.com"
	password = "wonderland"
	err = CreateCustomerCredentials(email, password)
	if err != nil {
		t.Fatal(err)
	}

	// ! ----------------------------------------------------------------------
	// ~ due to non unique mails creating multiple creds with same mail is allowed now
	// ! ----------------------------------------------------------------------

	// // Try to create credentials for already taken email.
	// // This should fail
	// email = "alice@bar.com"
	// password = "wonderland"
	// err = CreateCustomerCredentials(email, password)
	// if err == nil {
	// 	t.Fatal(err)
	// }

	// Change email
	err = ChangeEmail("foo@bar.com", "trent@bar.com")
	if err != nil {
		t.Fatal(err)
	}
	// Try to change email that does not exist.
	err = ChangeEmail("idont@exist.com", "a@b.com")
	if err == nil {
		t.Fatal(err)
	}

	// Try to change email with incorrect password
	err = ChangePassword("alice@bar.com", "wrong", "myNewPassWord", false)
	if err == nil {
		t.Fatal(err)
	}
	// Try to change email with correct password
	err = ChangePassword("alice@bar.com", "wonderland", "myNewPassWord", false)
	if err != nil {
		t.Fatal(err)
	}
	// Try new Password
	auth, err := CheckLoginCredentials("alice@bar.com", "myNewPassWord")
	if !auth || err != nil {
		t.Fatal(err)
	}

	// Delete Credentials of trent@bar.com
	err = DeleteCredential("trent@bar.com")
}

func TestCredentialsForGuestCustomer(t *testing.T) {

	test_utils.DropAllCollections()
	// Create credentials for a user
	email := "foo@bar.com"
	password := ""
	err := CreateCustomerCredentials(email, password)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := CheckLoginCredentials(email, password)
	if auth || err != nil {
		t.Fatal(err)
	}

}
