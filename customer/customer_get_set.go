package customer

import (
	"errors"
	"time"

	"github.com/foomo/shop/crypto"
	"github.com/foomo/shop/utils"
)

//------------------------------------------------------------------
// ~ PUBIC GETTERS
//------------------------------------------------------------------

func (customer *Customer) GetID() string {
	return customer.Id
}

func (address *Address) GetID() string {
	return address.Id
}

func (customer *Customer) GetEmail() string {
	return customer.Email
}

func (customer *Customer) GetPerson() *Person {
	return customer.Person
}

func (customer *Customer) GetCompany() *Company {
	return customer.Company
}

func (customer *Customer) GetLocalization() *Localization {
	return customer.Localization
}

func (customer *Customer) GetAddresses() []*Address {
	return customer.Addresses
}

func (customer *Customer) GetCreatedAt() time.Time {
	return customer.CreatedAt
}
func (customer *Customer) GetLastModifiedAt() time.Time {
	return customer.LastModifiedAt
}
func (customer *Customer) GetCreatedAtFormatted() string {
	return utils.GetFormattedTime(customer.CreatedAt)
}
func (customer *Customer) GetLastModifiedAtFormatted() string {
	return utils.GetFormattedTime(customer.LastModifiedAt)
}

func (customer *Customer) GetCrypto() *crypto.Crypto {
	return customer.Crypto
}

func (customer *Customer) GetAddress(id string) (*Address, error) {
	for _, address := range customer.Addresses {
		if address.Id == id {
			return address, nil
		}
	}
	return nil, errors.New("Could not find Address for id: " + id)
}

// GetDefaultShippingAddress returns the default shipping address if available, else returns first address
func (customer *Customer) GetDefaultShippingAddress() *Address {
	if len(customer.Addresses) == 0 {
		return nil
	}
	for _, address := range customer.Addresses {
		if address.IsDefaultShippingAddress {
			return address
		}
	}
	return customer.Addresses[0]
}

// GetDefaultBillingAddress returns the default billing address if available, else returns first address
func (customer *Customer) GetDefaultBillingAddress() *Address {
	if len(customer.Addresses) == 0 {
		return nil
	}
	for _, address := range customer.Addresses {
		if address.IsDefaultBillingAddress {
			return address
		}
	}
	return customer.Addresses[0]
}

func (c *Contacts) GetPrimaryContact() string {
	switch c.Primary {
	case ContactTypePhoneLandline:
		return string(ContactTypePhoneLandline) + ": " + c.PhoneLandLine
	case ContactTypePhoneMobile:
		return string(ContactTypePhoneMobile) + ": " + c.PhoneMobile
	case ContactTypeEmail:
		return string(ContactTypeEmail) + ": " + c.Email
	case ContactTypeSkype:
		return string(ContactTypeSkype) + ": " + c.Skype
	case ContactTypeFax:
		return string(ContactTypeFax) + ": " + c.Fax
	}
	return "No primary contact available!"
}

//------------------------------------------------------------------
// ~ PUBIC SETTERS
//------------------------------------------------------------------

// SetPassword stores a cryptographic hash and salt for password
func (customer *Customer) SetPassword(password string) error {
	crypto, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	customer.Crypto = crypto
	return customer.Upsert()
}
func (customer *Customer) SetDefaultShippingAddress(id string) error {
	for _, address := range customer.Addresses {
		if address.Id == id {
			address.IsDefaultShippingAddress = true
		} else {
			address.IsDefaultShippingAddress = false
		}
	}
	return customer.Upsert()
}
func (customer *Customer) SetDefaultBillingAddress(id string) error {
	for _, address := range customer.Addresses {
		if address.Id == id {
			address.IsDefaultBillingAddress = true
		} else {
			address.IsDefaultBillingAddress = false
		}
	}
	return customer.Upsert()
}
func (customer *Customer) SetModified() error {
	customer.LastModifiedAt = utils.TimeNow()
	return customer.Upsert()
}
func (customer *Customer) SetCompany(company *Company) error {
	customer.Company = company
	return customer.Upsert()
}
func (customer *Customer) SetLocalization(localization *Localization) error {
	customer.Localization = localization
	return customer.Upsert()
}
func (customer *Customer) SetPerson(person *Person) error {
	customer.Person = person
	return customer.Upsert()
}
func (customer *Customer) SetEmail(email string) error {
	customer.Email = email
	return customer.Upsert()
}
