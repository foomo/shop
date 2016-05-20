package customer

import (
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

type Address struct {
	Id                       string // is automatically set on AddAddress()
	Person                   *Person
	IsDefaultBillingAddress  bool
	IsDefaultShippingAddress bool
	Street                   string
	StreetNumber             string
	ZIP                      string
	City                     string
	Country                  string
	Company                  string      `bson:",omitempty"`
	Department               string      `bson:",omitempty"`
	Building                 string      `bson:",omitempty"`
	PostOfficeBox            string      `bson:",omitempty"`
	Custom                   interface{} `bson:",omitempty"`
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

func NewCustomer(customProvider CustomerCustomProvider) (*Customer, error) {
	customer := &Customer{
		Version: &history.Version{
			Number:    0,
			TimeStamp: time.Now(),
		},
		Id: unique.GetNewID(),
		//Id:             "mockIdCust",
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
	err := customer.Insert()

	// Retrieve customer again from. (Otherwise upserts on customer would fail because of missing mongo ObjectID)
	customer, err = GetCustomerById(customer.Id, customProvider) // TODO do not ignore this error
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

func (customer *Customer) Insert() error {
	return InsertCustomer(customer) // calls the method defined in persistor.go
}

func (customer *Customer) Upsert() error {
	return UpsertCustomer(customer) // calls the method defined in persistor.go
}
func (customer *Customer) Delete() error {
	return DeleteCustomer(customer)
}

// func (address *Address) OverrideId(id string) {
// 	address.id = id
// }

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

func CheckAccountAvailability(email string) bool {
	return false
}
