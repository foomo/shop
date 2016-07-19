package customer

import (
	"errors"
	"log"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/shop_error"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

const (
	ContactTypePhoneLandline ContactType    = "landline"
	ContactTypePhoneMobile   ContactType    = "mobile"
	ContactTypeEmail         ContactType    = "email"
	ContactTypeSkype         ContactType    = "skype"
	ContactTypeFax           ContactType    = "fax"
	SalutationTypeMr         SalutationType = "Mr"
	SalutationTypeMrs        SalutationType = "Mrs"
	SalutationTypeMrAndMrs   SalutationType = "MrAndMrs"
	SalutationTypeCompany    SalutationType = "Company" // TODO: find better wording
	SalutationTypeFamily     SalutationType = "Family"  // TODO: find better wording
	TitleTypeDr              TitleType      = "Dr"
	TitleTypeProf            TitleType      = "Prof."
	TitleTypeProfDr          TitleType      = "Prof. Dr."
	TitleTypePriest          TitleType      = "Priest" // TODO: find better wording
	CountryCodeGermany       CountryCode    = "DE"
	CountryCodeSwitzerland   CountryCode    = "CH"
	LanguageCodeGermany      LanguageCode   = "DE"
	LanguageCodeSwitzerland  LanguageCode   = "CH"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type ContactType string
type SalutationType string
type TitleType string
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
	Person         *Person
	IsGuest        bool
	Company        *Company
	Addresses      []*Address
	Localization   *Localization
	TacAgree       bool // Terms and Conditions
	Custom         interface{}
}

type Flags struct {
	forceUpsert bool // if true, Upsert is performed even if there is a version conflict. This is important for rollbacks.
}

type Contacts struct {
	PhoneLandLine string
	PhoneMobile   string
	Email         string
	Skype         string
	Fax           string
	Primary       ContactType
}

// Person is a field Customer and of Address
// Only Customer->Person has Contacts
type Person struct {
	FirstName  string
	MiddleName string
	LastName   string
	Title      TitleType
	Salutation SalutationType
	Birthday   string
	Contacts   *Contacts
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
	Fields() *bson.M
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
	if email == "" {
		return nil, errors.New(shop_error.ErrorRequiredFieldMissing)
	}
	// Check is desired Email is available
	available, err := CheckLoginAvailable(email)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New(shop_error.ErrorNotFound + " Login " + email + " is already taken!")
	}

	err = CreateCustomerCredentials(email, password)
	if err != nil {
		return nil, err
	}

	customer := &Customer{
		Flags:          &Flags{},
		Version:        version.NewVersion(),
		Id:             unique.GetNewID(),
		Email:          lc(email),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person: &Person{
			Contacts: &Contacts{},
		},
		Localization: &Localization{},
	}

	if password == "" {
		customer.IsGuest = true
	}

	if customProvider != nil {
		customer.Custom = customProvider.NewCustomerCustom()
	}
	// Store order in database
	err = customer.insert()

	// Retrieve customer again from. (Otherwise upserts on customer would fail because of missing mongo ObjectID)
	customer, err = GetCustomerById(customer.Id, customProvider)
	if err != nil {
		return nil, err
	}
	return customer, err
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON CUSTOMER
//------------------------------------------------------------------

func (customer *Customer) ChangeEmail(email, newEmail string) error {
	err := ChangeEmail(email, newEmail)
	if err != nil {
		return err
	}
	customer.Email = lc(email)
	return customer.Upsert()
}
func (customer *Customer) ChangePassword(password, passwordNew string, force bool) error {
	err := ChangePassword(customer.Email, password, passwordNew, force)
	if err != nil {
		return err
	}
	return customer.Upsert()
}

// Unlinks order from database
// After unlink, persistent changes on order are no longer possible until it is retrieved again from db.
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

// AddAddress adds a new address to the customers profile and returns its unique id
func (customer *Customer) AddAddress(address *Address) (string, error) {
	if address.Person == nil {
		return "", errors.New(shop_error.ErrorRequiredFieldMissing)
	}

	// Return error if required field is missing
	if address.Person.Salutation == "" || address.Person.FirstName == "" || address.Person.LastName == "" || address.Street == "" || address.StreetNumber == "" || address.ZIP == "" || address.City == "" || address.Country == "" {
		return "", errors.New(shop_error.ErrorRequiredFieldMissing)
	}

	// Create a unique id for this address
	address.Id = unique.GetNewID()
	// Prevent nil pointer in case we get an incomplete address
	if address.Person == nil {
		address.Person = &Person{
			Contacts: &Contacts{},
		}
	} else if address.Person.Contacts == nil {
		address.Person.Contacts = &Contacts{}
	}

	// If Person of Customer is still empty and this is the first address
	// added to the customer, Person of Address is adopted for Customer
	if len(customer.Addresses) == 0 && customer.Person.LastName == "" {
		*customer.Person = *address.Person
	}

	customer.Addresses = append(customer.Addresses, address)
	// Set Address as primary if this is the first added address
	if address.IsPrimary || len(customer.Addresses) == 1 {
		err := customer.SetPrimaryAddress(address.Id)
		if err != nil {
			return address.Id, err
		}
	}

	// If this is the first added Address, it's set as billing address
	if address.IsDefaultBillingAddress || len(customer.Addresses) == 1 {
		err := customer.SetDefaultBillingAddress(address.Id)
		if err != nil {
			return address.Id, err
		}
	}
	// If this is the first added Address, it's set as shipping address
	if address.IsDefaultShippingAddress || len(customer.Addresses) == 1 {
		err := customer.SetDefaultShippingAddress(address.Id)
		if err != nil {
			return address.Id, err
		}
	}
	return address.Id, customer.Upsert()
}
func (customer *Customer) RemoveAddress(id string) {
	for index, address := range customer.Addresses {
		if address.Id == id {
			customer.Addresses = append(customer.Addresses[:index], customer.Addresses[index+1:]...)
		}
	}
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
