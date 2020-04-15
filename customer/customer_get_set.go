package customer

import (
	"errors"
	"time"

	"github.com/foomo/shop/address"
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

func (customer *Customer) GetPerson() *address.Person {
	return customer.Person
}

func (customer *Customer) GetCompany() *Company {
	return customer.Company
}

func (customer *Customer) GetLocalization() *Localization {
	return customer.Localization
}

func (customer *Customer) GetAddresses() []*address.Address {
	return customer.Addresses
}

// GetSecondaryAddress returns all Addresses but the default billing address (Which is the address the customer is mainly associated with)
func (customer *Customer) GetSecondaryAddresses() []*address.Address {

	addresses := []*address.Address{}
	for _, addr := range customer.GetAddresses() {
		if addr.Type != address.AddressDefaultBilling {
			addresses = append(addresses, addr)
		}
	}
	return addresses
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

func (customer *Customer) GetAddress(id string) (*address.Address, error) {
	for _, address := range customer.Addresses {
		if address.Id == id {
			return address, nil
		}
	}
	return nil, errors.New(shop_error.ErrorNotFound + "Could not find Address for id: " + id)
}

func (customer *Customer) GetAddressById(id string) (*address.Address, error) {
	for _, address := range customer.GetAddresses() {
		if address.GetID() == id {
			return address, nil
		}
	}
	return nil, errors.New(shop_error.ErrorNotFound)
}

func (customer *Customer) GetScoreForAddress(addressId string) (*address.Score, error) {
	address, err := customer.GetAddressById(addressId)
	if err != nil {
		return nil, err
	}
	if address.Score == nil {
		return nil, errors.New("Score for address with id" + addressId + "is nil.")
	}
	return address.Score, nil
}

// GetPrimaryAddress returns the default billing address
func (customer *Customer) GetPrimaryAddress() (*address.Address, error) {
	return customer.GetDefaultBillingAddress()
}

// GetDefaultShippingAddress returns the default shipping address if available, else returns first address
func (customer *Customer) GetDefaultShippingAddress() (*address.Address, error) {
	if len(customer.Addresses) == 0 {
		return nil, errors.New(shop_error.ErrorNotFound + " Customer does not have an address")
	}
	for _, addr := range customer.Addresses {
		if addr.Type == address.AddressDefaultShipping {
			return addr, nil
		}
	}
	return customer.Addresses[0], nil
}

func (customer *Customer) GetDefaultBillingAddressID() (string, error) {
	address, err := customer.GetDefaultBillingAddress()
	if err != nil {
		return "", err
	}
	return address.GetID(), nil

}
func (customer *Customer) GetDefaultShippingAddressID() (string, error) {
	address, err := customer.GetDefaultShippingAddress()
	if err != nil {
		return "", err
	}
	return address.GetID(), nil

}

// GetDefaultBillingAddress returns the default billing address if available, else returns first address
func (customer *Customer) GetDefaultBillingAddress() (*address.Address, error) {
	if len(customer.Addresses) == 0 {
		return nil, errors.New(shop_error.ErrorNotFound + " Customer does not have an address")
	}
	for _, addr := range customer.Addresses {
		if addr.Type == address.AddressDefaultBilling {
			return addr, nil
		}
	}
	return customer.Addresses[0], nil
}

//------------------------------------------------------------------
// ~ PUBLIC SETTERS
//------------------------------------------------------------------

func (customer *Customer) SetDefaultShippingAddress(id string) error {
	for _, addr := range customer.Addresses {
		if addr.Id == id {
			addr.Type = address.AddressDefaultShipping
		} else {
			if addr.Type == address.AddressDefaultShipping {
				addr.Type = address.AddressOther
			}
		}
	}
	return customer.Upsert()
}
func (customer *Customer) SetDefaultBillingAddress(id string) error {
	for _, addr := range customer.Addresses {
		if addr.Id == id {
			addr.Type = address.AddressDefaultBilling
		} else {
			if addr.Type == address.AddressDefaultBilling {
				addr.Type = address.AddressOther
			}
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
func (customer *Customer) SetPerson(person *address.Person) error {
	customer.Person = person
	return customer.Upsert()
}
