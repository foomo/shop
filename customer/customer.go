package customer

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
	AddressTypeDelivery      AddressType    = "delivery"
	AddressTypeBilling       AddressType    = "billing"
	CountryCodeGermany       CountryCode    = "DE"
	CountryCodeSwistzerland  CountryCode    = "CH"
	LanguageCodeGermany      LanguageCode   = "DE"
	LanguageCodeSwistzerland LanguageCode   = "CH"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC TYPES
+++++++++++++++++++++++++++++++++++++++++++++++++ */

type AddressType string
type ContactType string
type SalutationType string
type TitleType string
type LanguageCode string
type CountryCode string

// TOP LEVEL OBJECT
type Customer struct {
	CustomerID      string
	Person          *Person
	Company         *Company `bson:",omitempty"`
	AddressBilling  *Address
	AddressShipping *Address

	Localization *Localization
	Custom       interface{} `bson:",omitempty"`
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
	//Address    []*Address
}

type Company struct {
	Name string
	Type string
}

type Address struct {
	Person        *Person
	Type          AddressType // e.g. Shipping or Billing Address
	Street        string
	StreetNumber  string
	ZIP           string
	City          string
	Country       string
	PostOfficeBox string      `bson:",omitempty"`
	Custom        interface{} `bson:",omitempty"`
}

type Localization struct {
	LanguageCode LanguageCode
	CountryCode  CountryCode
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

func (customer *Customer) GetAddresses() []*Address {
	addresses := []*Address{}
	addresses = append(addresses, customer.AddressBilling)
	addresses = append(addresses, customer.AddressShipping)
	return addresses
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
