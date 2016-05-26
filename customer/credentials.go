package customer

import (
	"errors"

	"github.com/foomo/shop/crypto"
	"github.com/foomo/shop/history"
	"gopkg.in/mgo.v2/bson"
)

type CustomerCredentials struct {
	BsonId  bson.ObjectId `bson:"_id,omitempty"`
	Version *history.Version
	Email   string
	Crypto  *crypto.Crypto
}

// GetCredentials from db
func GetCredentials(email string) (*CustomerCredentials, error) {
	p := GetCredentialsPersistor()
	credentials := &CustomerCredentials{}
	err := p.GetCollection().Find(&bson.M{"email": email}).One(credentials)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

// CreateCustomerCredentials
func CreateCustomerCredentials(email, password string) error {
	crypto, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	credentials := &CustomerCredentials{
		Version: history.NewVersion(),
		Email:   email,
		Crypto:  crypto,
	}
	p := GetCredentialsPersistor()
	return p.GetCollection().Insert(credentials)

}

// CheckLoginAvailable returns true if the email address is available as login credential
func CheckLoginAvailable(email string) (bool, error) {
	p := GetCustomerPersistor()
	query := p.GetCollection().Find(&bson.M{"email": email})
	count, err := query.Count()
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// CheckLoginCredentials returns true if  customer with email exists and password matches with the hash stores in customers Crypto
func CheckLoginCredentials(email, password string) (bool, error) {
	credentials, err := GetCredentials(email)
	if err != nil {
		return false, err
	}
	return crypto.VerifyPassword(credentials.Crypto, password), nil
}

// ChangePassword changes the password of the user.
// If force, passworldOld is irrelevant and the password is changed in any case.
func (customer *Customer) ChangePassword(email, password, passwordNew string, force bool) error {
	credentials, err := GetCredentials(email)
	if err != nil {
		return err
	}

	auth := force || crypto.VerifyPassword(credentials.Crypto, password)
	if auth {
		newCrypto, err := crypto.HashPassword(passwordNew)
		if err != nil {
			return err
		}
		credentials.Crypto = newCrypto
		credentials.Version.Increment()
		_, err = GetCredentialsPersistor().GetCollection().UpsertId(credentials.BsonId, credentials)
		return err
	}

	return errors.New("Authorization Error: Could not change password.")
}

func ChangeEmail(email, newEmail string) error {
	available, err := CheckLoginAvailable(newEmail)
	if err != nil {
		return err
	}
	if !available {
		return errors.New("Could not change Email: \"" + newEmail + "\" is already taken!")
	}
	credentials, err := GetCredentials(email)
	if err != nil {
		return err
	}
	credentials.Email = newEmail
	credentials.Version.Increment()
	_, err = GetCredentialsPersistor().GetCollection().UpsertId(credentials.BsonId, credentials)
	return err

}
