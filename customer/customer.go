package customer

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/foomo/shop/address"
	"github.com/foomo/shop/shop_error"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

const (
	CountryCodeGermany       CountryCode = "DE"
	CountryCodeFrance        CountryCode = "FR"
	CountryCodeSwitzerland   CountryCode = "CH"
	CountryCodeAustria       CountryCode = "AT"
	CountryCodeLiechtenstein CountryCode = "LI"
	CountryCodeItaly         CountryCode = "IT"

	LanguageCodeFrance      LanguageCode = "fr"
	LanguageCodeGermany     LanguageCode = "de"
	LanguageCodeSwitzerland LanguageCode = "ch"

	KeyAddrKey     = "addrkey"
	KeyAddrKeyHash = "addrkeyhash"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type (
	LanguageCode string
	CountryCode  string
)

// TOP LEVEL OBJECT
// private, so that changes are limited by API
type Customer struct {
	BsonId         bson.ObjectId `bson:"_id,omitempty"`
	AddrKey        string        // unique id which will replace Id for primary way of retrieval
	AddrKeyHash    string        // unique id which will replace Id for primary way of retrieval
	ExternalID     string
	Id             string
	unlinkDB       bool // if true, changes to Customer are not stored in database
	Version        *version.Version
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Email          string // unique, used as Login Credential
	Person         *address.Person
	IsGuest        bool
	Company        *Company
	Addresses      []*address.Address
	Localization   *Localization
	TacAgree       bool // Terms and Conditions
	Custom         interface{}
}

type Company struct {
	Name string
	Type string
}

type Localization struct {
	LanguageCode LanguageCode
	CountryCode  CountryCode
}

type CustomerCustomProvider interface {
	NewCustomerCustom() interface{}
	NewAddressCustom() interface{}
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewCustomer creates a new customer in the database and returns it.
// addrkey must be unique for a customer.
// customerProvider is required.
// mailContact can be nil, if not nil a valid email address is required.
func NewCustomer(addrkey string, addrkeyHash string, externalID string, mailContact *address.Contact, customProvider CustomerCustomProvider) (*Customer, error) {
	var mErr *multierror.Error

	if addrkey == "" {
		mErr = multierror.Append(mErr, errors.New("required addrkey is empty"))
	}
	if addrkeyHash == "" {
		mErr = multierror.Append(mErr, errors.New("required addrkeyHash is empty"))
	}
	if externalID == "" {
		mErr = multierror.Append(mErr, errors.New("required externalID is empty"))
	}
	if mailContact != nil {
		if !strings.ContainsRune(mailContact.Value, '@') {
			mErr = multierror.Append(mErr, errors.New("required email address in mailContact.Value is empty"))
		}
		if !mailContact.IsMail() {
			mErr = multierror.Append(mErr, fmt.Errorf("required mailContact must have string type %q", address.ContactTypeEmail))
		}
	}
	if customProvider == nil {
		mErr = multierror.Append(mErr, errors.New("custom provider not set"))
	}

	if mErr.ErrorOrNil() != nil {
		return nil, mErr.ErrorOrNil()
	}

	var person *address.Person
	var email string
	if mailContact != nil {
		email = lc(mailContact.Value)
		person = &address.Person{
			Contacts: map[string]*address.Contact{
				mailContact.ID: mailContact,
			},
			DefaultContacts: map[address.ContactType]string{
				address.ContactTypeEmail: mailContact.ID,
			},
		}
	}
	customer := &Customer{
		Version:        version.NewVersion(),
		Id:             unique.GetNewID(),
		ExternalID:     externalID,
		AddrKey:        addrkey,
		AddrKeyHash:    addrkeyHash,
		Email:          email,
		IsGuest:        false,
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person:         person,
		Localization:   &Localization{},
		Custom:         customProvider.NewCustomerCustom(),
	}

	// initial version should be 1
	customer.Version.Increment()

	// persist customer in database
	if err := customer.insert(); err != nil {
		if mgo.IsDup(err) {
			return nil, shop_error.ErrorDuplicateKey
		}
		return nil, err
	}

	// retrieve customer again from database,
	// otherwise upserts on customer would fail because of missing mongo ObjectID)
	return GetCustomerById(customer.Id, customProvider)
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON CUSTOMER
//------------------------------------------------------------------

func (customer *Customer) insert() error {
	return insertCustomer(customer)
}

func (customer *Customer) Upsert() error {
	return UpsertCustomer(customer)
}

func (customer *Customer) UpsertAndGetCustomer(customProvider CustomerCustomProvider) (*Customer, error) {
	return UpsertAndGetCustomer(customer, customProvider)
}

func (customer *Customer) Delete() error {
	return DeleteCustomer(customer)
}

func (customer *Customer) OverrideId(id string) error {
	customer.Id = id
	return customer.Upsert()
}

func (customer *Customer) AddDefaultBillingAddress(addr *address.Address) (string, error) {
	addr.Type = address.AddressDefaultBilling
	return customer.AddAddress(addr)
}

func (customer *Customer) AddDefaultShippingAddress(addr *address.Address) (string, error) {
	addr.Type = address.AddressDefaultShipping
	return customer.AddAddress(addr)
}

// AddAddress adds a new address to the customers profile and returns its unique id
func (customer *Customer) AddAddress(addr *address.Address) (string, error) {
	err := addr.IsComplete()
	if err != nil {
		log.Println("Error", err)
		return "", err
	}
	// Create a unique id for this address
	addr.Id = unique.GetNewID()
	// Prevent nil pointer in case we get an incomplete address
	if addr.Person == nil {
		addr.Person = &address.Person{}
	}
	if addr.Person.Contacts == nil {
		addr.Person.Contacts = map[string]*address.Contact{}
		addr.Person.DefaultContacts = map[address.ContactType]string{}
	}

	// If Person of Customer is still empty and this is the first address
	// added to the customer, Person of Address is adopted for Customer
	// log.Println("is customer nil: ", customer == nil)
	// log.Println("is customer.person nil: ", customer.Person == nil)
	if customer.Person == nil {
		log.Println("WARNING: customer.Person must not be nil: customerID: " + customer.GetID() + ", AddressID: " + addr.Id)
		customer.Person = &address.Person{
			Contacts:        map[string]*address.Contact{},
			DefaultContacts: map[address.ContactType]string{},
		}
		*customer.Person = *addr.Person
	} else if len(customer.Addresses) == 0 && customer.Person != nil && customer.Person.LastName == "" {
		*customer.Person = *addr.Person
	}

	customer.Addresses = append(customer.Addresses, addr)

	// If this is the first added Address, it's set as billing address
	if addr.Type == address.AddressDefaultBilling || len(customer.Addresses) == 1 {
		err := customer.SetDefaultBillingAddress(addr.Id)
		if err != nil {
			return addr.Id, err
		}
	}
	// If this is the first added Address, it's set as shipping address
	if addr.Type == address.AddressDefaultShipping {
		err := customer.SetDefaultShippingAddress(addr.Id)
		if err != nil {
			return addr.Id, err
		}
	}
	return addr.Id, customer.Upsert()
}

func (customer *Customer) RemoveAddress(id string) error {
	addresses := []*address.Address{}
	for _, address := range customer.Addresses {
		if address.Id == id {
			continue
		}
		addresses = append(addresses, address)
	}
	customer.Addresses = addresses
	return customer.Upsert()
}

func (customer *Customer) ChangeAddress(addr *address.Address) error {
	addressToBeChanged, err := customer.GetAddressById(addr.GetID())
	if err != nil {
		log.Println("Error: Could not find address with id "+addr.GetID(), "for customer ", customer.Person.LastName)
		return err
	}
	err = addr.IsComplete()
	if err != nil {
		return err
	}
	*addressToBeChanged = *addr
	*addressToBeChanged.Person = *addr.Person
	return customer.Upsert()
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// lc returns lowercase version of string
func lc(s string) string {
	return strings.ToLower(s)
}
