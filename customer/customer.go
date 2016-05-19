package customer

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/crypto"
	"github.com/foomo/shop/event_log"
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
	BsonID         bson.ObjectId `bson:"_id,omitempty"`
	Id             string
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

func NewCustomer(custom interface{}) *Customer {
	customer := &Customer{
		Id:             unique.GetNewID(),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Person: &Person{
			Contacts: &Contacts{},
		},
		Localization: &Localization{},
		Custom:       custom,
	}

	return customer
}

func (address *Address) GetID() string {
	return address.Id
}

func (customer *Customer) Insert() error {
	return InsertCustomer(customer) // calls the method defined in persistor.go
}

func (customer *Customer) Upsert() error {
	return UpsertCustomer(customer) // calls the method defined in persistor.go
}
func (customer *Customer) Delete() error {
	return nil // TODO delete order in db
}

// func (address *Address) OverrideId(id string) {
// 	address.id = id
// }

func (customer *Customer) GetID() string {
	return customer.Id
}

func (customer *Customer) OverrideId(id string) {
	customer.Id = id
}

func (customer *Customer) GetEmail() string {
	return customer.Email
}

func (customer *Customer) SetEmail(email string) {
	customer.Email = email
}
func (customer *Customer) GetPassword() string {
	return customer.Email
}

func (customer *Customer) SetPassword(password string) {
	customer.Password = password
}
func (customer *Customer) GetPerson() *Person {
	return customer.Person
}

func (customer *Customer) SetPerson(person *Person) {
	customer.Person = person
}
func (customer *Customer) GetCompany() *Company {
	return customer.Company
}

func (customer *Customer) SetCompany(company *Company) {
	customer.Company = company
}
func (customer *Customer) GetLocalization() *Localization {
	return customer.Localization
}

func (customer *Customer) SetLocalization(localization *Localization) {
	customer.Localization = localization
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

func (customer *Customer) SetDefaultShippingAddress(id string) {
	for _, address := range customer.Addresses {
		if address.Id == id {
			address.IsDefaultShippingAddress = true
		} else {
			address.IsDefaultShippingAddress = false
		}
	}
}
func (customer *Customer) SetDefaultBillingAddress(id string) {
	for _, address := range customer.Addresses {
		if address.Id == id {
			address.IsDefaultBillingAddress = true
		} else {
			address.IsDefaultBillingAddress = false
		}
	}
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
func (customer *Customer) GetAddress(id string) (*Address, error) {
	for _, address := range customer.Addresses {
		if address.Id == id {
			return address, nil
		}
	}
	return nil, errors.New("Could not find Address for id: " + id)
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

func (customer *Customer) SetModified() {
	customer.LastModifiedAt = utils.TimeNow()
}
