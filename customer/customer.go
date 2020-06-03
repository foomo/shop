package customer

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/foomo/shop/address"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
	"github.com/hashicorp/go-multierror"
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
	Flags          *Flags
	Version        *version.Version
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Email          string // unique, used as Login Credential
	Person         *address.Person
	IsGuest        bool
	IsComplete     bool // upsert checks if customer data is complete, so that a customer could place an order
	Company        *Company
	Addresses      []*address.Address
	Localization   *Localization
	TacAgree       bool // Terms and Conditions
	Custom         interface{}
}

type Flags struct {
	forceUpsert bool // if true, Upsert is performed even if there is a version conflict. This is important for rollbacks.
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
// addrkey must be unique for a customer
// customerProvider may be nil at this point.
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
	if mailContact == nil {
		mErr = multierror.Append(mErr, errors.New("required mailContact is empty"))
	} else {
		if mailContact.Value == "" {
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

	email := lc(mailContact.Value)
	customer := &Customer{
		Flags:          &Flags{},
		Version:        version.NewVersion(),
		Id:             unique.GetNewID(),
		ExternalID:     externalID,
		AddrKey:        addrkey,
		AddrKeyHash:    addrkeyHash,
		Email:          email,
		IsGuest:        false,
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person: &address.Person{
			Contacts: map[string]*address.Contact{
				mailContact.ID: mailContact,
			},
			DefaultContacts: map[address.ContactType]string{
				address.ContactTypeEmail: mailContact.ID,
			},
		},
		Localization: &Localization{},
		Custom:       customProvider.NewCustomerCustom(),
	}

	// persist customer in database
	errInsert := customer.insert()
	if errInsert != nil {
		return nil, errInsert
	}

	// retrieve customer again from database,
	// otherwise upserts on customer would fail because of missing mongo ObjectID)
	return GetCustomerById(customer.Id, customProvider)
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON CUSTOMER
//------------------------------------------------------------------

func (customer *Customer) ChangeEmail(email, newEmail string) error {
	// lower case
	email = lc(email)
	newEmail = lc(newEmail)

	customer.Email = newEmail
	for _, addr := range customer.GetAddresses() {
		for _, contact := range addr.Person.Contacts {
			if contact.IsMail() && contact.Value == email {
				contact.Value = newEmail
			}
		}
	}
	return customer.Upsert()
}

func (customer *Customer) ChangePassword(password, passwordNew string, force bool) error {
	err := ChangePassword(customer.Email, password, passwordNew, force)
	if err != nil {
		return err
	}
	return customer.Upsert()
}

// Unlinks customer from database
// After unlink, persistent changes on customer are no longer possible until it is retrieved again from db.
func (customer *Customer) UnlinkFromDB() {
	customer.unlinkDB = true
}

func (customer *Customer) LinkDB() {
	customer.unlinkDB = false
}

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

func (customer *Customer) Rollback(version int) error {
	return Rollback(customer.GetID(), version)
}

func (customer *Customer) OverrideId(id string) error {
	customer.Id = id
	return customer.Upsert()
}

// checkFields Checks if all required fields are specified
// @TODO which are teh required fields
func CheckRequiredAddressFields(address *address.Address) error {
	// Return error if required field is missing
	var mErr *multierror.Error

	if address.Person == nil {
		mErr = multierror.Append(mErr, errors.New("required person is nil"))
	} else {
		if address.Person.Salutation == "" {
			mErr = multierror.Append(mErr, errors.New("required person salutation is empty"))
		}
		if address.Person.FirstName == "" {
			mErr = multierror.Append(mErr, errors.New("required person firstname is empty"))
		}
		if address.Person.LastName == "" {
			mErr = multierror.Append(mErr, errors.New("required person lastname is empty"))
		}
	}
	if address.Street == "" {
		mErr = multierror.Append(mErr, errors.New("required address street is empty"))
	}
	if address.StreetNumber == "" {
		mErr = multierror.Append(mErr, errors.New("required address street number is empty"))
	}
	if address.ZIP == "" {
		mErr = multierror.Append(mErr, errors.New("required address zip is empty"))
	}
	if address.City == "" {
		mErr = multierror.Append(mErr, errors.New("required address city is empty"))
	}
	if address.Country == "" {
		mErr = multierror.Append(mErr, errors.New("required address country is empty"))
	}

	return mErr.ErrorOrNil()
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
	err := CheckRequiredAddressFields(addr)
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
	err = CheckRequiredAddressFields(addr)
	if err != nil {
		return err
	}
	*addressToBeChanged = *addr
	*addressToBeChanged.Person = *addr.Person
	return customer.Upsert()
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// DiffTwoLatestCustomerVersions compares the two latest Versions of Customer found in version.
// If openInBrowser, the result is automatically displayed in the default browser.
func DiffTwoLatestCustomerVersions(customerId string, customProvider CustomerCustomProvider, openInBrowser bool) (string, error) {
	version, err := GetCurrentVersionOfCustomerFromVersionsHistory(customerId)
	if err != nil {
		return "", err
	}

	return DiffCustomerVersions(customerId, version.Current-1, version.Current, customProvider, openInBrowser)
}

func DiffCustomerVersions(customerId string, versionA int, versionB int, customProvider CustomerCustomProvider, openInBrowser bool) (string, error) {
	if versionA <= 0 || versionB <= 0 {
		return "", errors.New("Error: Version must be greater than 0")
	}
	name := "customer_v" + strconv.Itoa(versionA) + "_vs_v" + strconv.Itoa(versionB)
	customerVersionA, err := GetCustomerByVersion(customerId, versionA, customProvider)
	if err != nil {
		return "", err
	}
	customerVersionB, err := GetCustomerByVersion(customerId, versionB, customProvider)
	if err != nil {
		return "", err
	}

	html, err := version.DiffVersions(customerVersionA, customerVersionB)
	if err != nil {
		return "", err
	}
	if openInBrowser {
		err := utils.OpenInBrowser(name, html)
		if err != nil {
			log.Println(err)
		}
	}
	return html, err
}
