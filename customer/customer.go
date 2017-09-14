package customer

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/foomo/shop/address"
	"github.com/foomo/shop/crypto"
	"github.com/foomo/shop/shop_error"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

const (
	CountryCodeGermany       CountryCode  = "DE"
	CountryCodeFrance        CountryCode  = "FR"
	CountryCodeSwitzerland   CountryCode  = "CH"
	CountryCodeLiechtenstein CountryCode  = "LI"
	LanguageCodeFrance       LanguageCode = "fr"
	LanguageCodeGermany      LanguageCode = "de"
	LanguageCodeSwitzerland  LanguageCode = "ch"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type LanguageCode string
type CountryCode string

// TOP LEVEL OBJECT
// private, so that changes are limited by API
type Customer struct {
	BsonId         bson.ObjectId `bson:"_id,omitempty"`
	Id             string        // Email is used as LoginID, but can change. This is never changes!
	unlinkDB       bool          // if true, changes to Customer are not stored in database
	Flags          *Flags
	Version        *version.Version
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Email          string // unique, used as Login Credential
	Person         *address.Person
	IsGuest        bool
	IsLoggedIn     bool
	Company        *Company
	Addresses      []*address.Address
	Localization   *Localization
	TacAgree       bool // Terms and Conditions
	Tracking       *Tracking
	Custom         interface{}
}

type Tracking struct {
	TrackingID string
	SessionIDs []string
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

// NewGuestCustomer creates a new Customer in the database and returns it.
// Per default, customers have an empty password which does not grant access at login.
// To transform Guests to regular Customers, the password must be changed and the customer must be marked as regular
func NewGuestCustomer(email string, customProvider CustomerCustomProvider) (*Customer, error) {
	return NewCustomer(email, "", customProvider)
}

// NewCustomer creates a new Customer in the database and returns it.
// Email must be unique for a customer. customerProvider may be nil at this point.
func NewCustomer(email, password string, customProvider CustomerCustomProvider) (*Customer, error) {
	log.Println("=== Creating new customer ", email)
	isGuest := password == ""
	if email == "" {
		return nil, errors.New(shop_error.ErrorRequiredFieldMissing)
	}
	//var err error
	// We only create credentials if a customer is available.
	// A guest customer gets a new entry in the customer db for each order!
	// if !isGuest {
	// 	// Check is desired Email is available
	// 	available, err := CheckLoginAvailable(email)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if !available {
	// 		return nil, errors.New(shop_error.ErrorNotFound + " Login " + email + " is already taken!")
	// 	}

	// 	// These credentials are not used at the moment
	// 	// err = CreateCustomerCredentials(email, password)
	// 	// if err != nil {
	// 	// 	return nil, err
	// 	// }
	// }

	customer := &Customer{
		Flags:          &Flags{},
		Version:        version.NewVersion(),
		Id:             unique.GetNewID(),
		Email:          utils.IteString(isGuest, "", lc(email)), // If Customer is a guest, we do not set the email address. This field should be unique in the database (and would not be if the guest ordered twice).
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person: &address.Person{
			Contacts: &address.Contacts{
				Email: email,
			},
		},
		Localization: &Localization{},
		Tracking:     &Tracking{},
	}

	trackingId, err := crypto.CreateHash(customer.GetID())
	if err != nil {
		return nil, err
	}
	customer.Tracking.TrackingID = "tid" + trackingId

	if isGuest {
		customer.IsGuest = true
	}

	if customProvider != nil {
		customer.Custom = customProvider.NewCustomerCustom()
	}
	// Store order in database
	err = customer.insert()
	if err != nil {
		log.Println("Could not insert customer", email)
		return nil, err
	}
	// Retrieve customer again from. (Otherwise upserts on customer would fail because of missing mongo ObjectID)
	return GetCustomerById(customer.Id, customProvider)
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON CUSTOMER
//------------------------------------------------------------------

func (customer *Customer) ChangeEmail(email, newEmail string) error {
	// err := ChangeEmail(email, newEmail)
	// if err != nil {
	// 	return err
	// }
	customer.Email = lc(newEmail)
	for _, addr := range customer.GetAddresses() {
		if addr.Person.Contacts.Email == lc(email) {
			addr.Person.Contacts.Email = lc(newEmail)
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
	if address.Person == nil || address.Person.Salutation == "" || address.Person.FirstName == "" || address.Person.LastName == "" || address.Street == "" || address.StreetNumber == "" || address.ZIP == "" || address.City == "" || address.Country == "" {
		return errors.New(shop_error.ErrorRequiredFieldMissing + "\n" + utils.ToJSON(address))
	}
	return nil
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
		addr.Person = &address.Person{
			Contacts: &address.Contacts{},
		}
	} else if addr.Person.Contacts == nil {
		addr.Person.Contacts = &address.Contacts{}
	}

	// If Person of Customer is still empty and this is the first address
	// added to the customer, Person of Address is adopted for Customer
	log.Println("is customer nil: ", customer == nil)
	log.Println("is customer.person nil: ", customer.Person == nil)
	if customer.Person == nil {
		log.Println("WARNING: customer.Person must not be nil: customerID: " + customer.GetID() + ", AddressID: " + addr.Id)
		customer.Person = &address.Person{
			Contacts: &address.Contacts{},
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
