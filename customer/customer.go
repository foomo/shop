package customer

import (
	"errors"
	"log"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/crypto"
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/history"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			CONSTANTS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

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

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC TYPES
+++++++++++++++++++++++++++++++++++++++++++++++++ */

type ContactType string
type SalutationType string
type TitleType string
type LanguageCode string
type CountryCode string

// TOP LEVEL OBJECT
// private, so that changes are limited by API
type Customer struct {
	Version        *history.Version
	unlinkDB       bool          // if true, changes to Customer are not stored in database
	BsonID         bson.ObjectId `bson:"_id,omitempty"`
	Id             string        // Email is used as LoginID. This is never changes!
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Email          string // Login Credential
	Crypto         *crypto.Crypto
	Person         *Person
	Company        *Company `bson:",omitempty"`
	Addresses      []*Address
	Localization   *Localization
	History        event_log.EventHistory
	Custom         interface{} `bson:",omitempty"`
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
	MiddleName string `bson:",omitempty"`
	LastName   string
	Title      TitleType `bson:",omitempty"`
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

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

// NewCustomer creates a new Customer in the database and returns it.
func NewCustomer(customProvider CustomerCustomProvider) (*Customer, error) {
	customer := &Customer{
		Version: &history.Version{
			Number:    0,
			TimeStamp: time.Now(),
		},
		Id:             unique.GetNewID(),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person: &Person{
			Contacts: &Contacts{},
		},
		Crypto:       &crypto.Crypto{},
		Localization: &Localization{},
		Custom:       customProvider.NewCustomerCustom(),
	}
	// Store order in database
	err := customer.insert()

	// Retrieve customer again from. (Otherwise upserts on customer would fail because of missing mongo ObjectID)
	customer, err = GetCustomerById(customer.Id, customProvider)
	if err != nil {
		return nil, err
	}
	return customer, err
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
	return insertCustomer(customer) // calls the method defined in persistor.go
}

func (customer *Customer) Upsert() error {

	return UpsertCustomer(customer) // calls the method defined in persistor.go
}
func (customer *Customer) Delete() error {
	return DeleteCustomer(customer)
}

func (customer *Customer) OverrideId(id string) {
	customer.Id = id
	customer.Upsert()
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

// DiffTwoLatestCustomerVersions compares the two latest Versions of Customer found in history.
// If openInBrowser, the result is automatically displayed in the default browser.
func DiffTwoLatestCustomerVersions(customerId string, customProvider CustomerCustomProvider, openInBrowser bool) (string, error) {
	version, err := GetCurrentVersionOfCustomerFromHistory(customerId)
	if err != nil {
		return "", err
	}

	return DiffCustomerVersions(customerId, version.Number-1, version.Number, customProvider, openInBrowser)
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
