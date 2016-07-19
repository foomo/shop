package customer

import (
	"errors"
	"time"

	"github.com/foomo/shop/shop_error"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
)

//------------------------------------------------------------------
// ~ PUBLIC GETTERS
//------------------------------------------------------------------

func (customer *Customer) GetID() string {
	return customer.Id
}

func (customer *Customer) GetVersion() *version.Version {
	return customer.Version
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

func (customer *Customer) GetAddress(id string) (*Address, error) {
	for _, address := range customer.Addresses {
		if address.Id == id {
			return address, nil
		}
	}
	return nil, errors.New(shop_error.ErrorNotFound + "Could not find Address for id: " + id)
}

func (customer *Customer) GetAddressById(id string) (*Address, error) {
	for _, address := range customer.GetAddresses() {
		if address.GetID() == id {
			return address, nil
		}
	}
	return nil, errors.New(shop_error.ErrorNotFound)
}

// GetDefaultShippingAddress returns the default shipping address if available, else returns first address
func (customer *Customer) GetPrimaryAddress() (*Address, error) {
	if len(customer.Addresses) == 0 {
		return nil, errors.New(shop_error.ErrorNotFound + " Customer does not have an address")
	}
	for _, address := range customer.Addresses {
		if address.IsPrimary {
			return address, nil
		}
	}
	// This case should not appear as first added address is set as primary
	return customer.Addresses[0], nil
}

// GetDefaultShippingAddress returns the default shipping address if available, else returns first address
func (customer *Customer) GetDefaultShippingAddress() (*Address, error) {
	if len(customer.Addresses) == 0 {
		return nil, errors.New(shop_error.ErrorNotFound + " Customer does not have an address")
	}
	for _, address := range customer.Addresses {
		if address.IsDefaultShippingAddress {
			return address, nil
		}
	}
	return customer.Addresses[0], nil
}

// GetDefaultBillingAddress returns the default billing address if available, else returns first address
func (customer *Customer) GetDefaultBillingAddress() (*Address, error) {
	if len(customer.Addresses) == 0 {
		return nil, errors.New(shop_error.ErrorNotFound + " Customer does not have an address")
	}
	for _, address := range customer.Addresses {
		if address.IsDefaultBillingAddress {
			return address, nil
		}
	}
	return customer.Addresses[0], nil
}

// GetPrimaryContact returns primary contact as string
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
// ~ PUBLIC SETTERS
//------------------------------------------------------------------

func (customer *Customer) SetPrimaryAddress(id string) error {
	for _, address := range customer.Addresses {
		if address.Id == id {
			address.IsPrimary = true
		} else {
			address.IsPrimary = false
		}
	}
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
