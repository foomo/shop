package customer

import (
	"errors"
	"log"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/history"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
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
	SalutationTypeMr         SalutationType = "Herr"
	SalutationTypeMrs        SalutationType = "Frau"
	TitleTypeDr              TitleType      = "Dr"
	TitleTypeProf            TitleType      = "Prof"
	TitleTypeProfDr          TitleType      = "ProfDr"
	CountryCodeGermany       CountryCode    = "DE"
	CountryCodeSwistzerland  CountryCode    = "CH"
	LanguageCodeGermany      LanguageCode   = "DE"
	LanguageCodeSwistzerland LanguageCode   = "CH"
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
	Version        *history.Version
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Email          string // unique, used as Login Credential
	Person         *Person
	Company        *Company
	Addresses      []*Address
	Localization   *Localization
	History        event_log.EventHistory
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
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// NewCustomer creates a new Customer in the database and returns it.
// Email must be unique for a customer. customerProvider may be nil at this point.
func NewCustomer(email, password string, customProvider CustomerCustomProvider) (*Customer, error) {

	// Check is desired Email is available
	available, err := CheckLoginAvailable(email)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("Login " + email + " is already taken!")
	}

	err = CreateCustomerCredentials(email, password)
	if err != nil {
		return nil, err
	}

	customer := &Customer{
		Flags:          &Flags{},
		Version:        history.NewVersion(),
		Id:             unique.GetNewID(),
		Email:          lc(email),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person: &Person{
			Contacts: &Contacts{},
		},
		Localization: &Localization{},
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

func (customer *Customer) AddAddress(address *Address) {

	address.Id = unique.GetNewID()
	customer.Addresses = append(customer.Addresses, address)
	// Adjust default addresses
	if address.IsDefaultBillingAddress {
		customer.SetDefaultBillingAddress(address.Id)
	}
	if address.IsDefaultShippingAddress {
		customer.SetDefaultShippingAddress(address.Id)
	}
}
func (customer *Customer) RemoveAddress(id string) {
	for index, address := range customer.Addresses {
		if address.Id == id {
			customer.Addresses = append(customer.Addresses[:index], customer.Addresses[index+1:]...)
		}
	}
}

// DiffTwoLatestCustomerVersions compares the two latest Versions of Customer found in history.
// If openInBrowser, the result is automatically displayed in the default browser.
func DiffTwoLatestCustomerVersions(customerId string, customProvider CustomerCustomProvider, openInBrowser bool) (string, error) {
	version, err := GetCurrentVersionOfCustomerFromHistory(customerId)
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

	html, err := history.DiffVersions(customerVersionA, customerVersionB)
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
