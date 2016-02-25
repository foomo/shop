package customer

type ContactType string

const (
	ContactTypePhoneLandline ContactType = "landline"
	ContactTypePhoneMobile   ContactType = "mobile"
	ContactTypeEmail         ContactType = "email"
	ContactTypeSkype         ContactType = "skype"
	ContactTypeFax           ContactType = "fax"
)

type SalutationType string

const (
	SalutationTypeMr  SalutationType = "Herr"
	SalutationTypeMrs SalutationType = "Frau"
)

type TitleType string

const (
	TitleTypeDr     TitleType = "Dr"
	TitleTypeProf   TitleType = "Prof"
	TitleTypeProfDr TitleType = "ProfDr"
)

// refactored from map because of validation
type Contacts struct {
	PhoneLandLine string
	PhoneMobile   string
	Email         string
	Skype         string
	Fax           string
	Primary       ContactType
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

type Person struct {
	FirstName  string
	MiddleName string `bson:",omitempty"`
	LastName   string
	Title      TitleType `bson:",omitempty"`
	Salutation SalutationType
	Contacts   *Contacts
}

type Company struct {
	Name     string
	Type     string
	Contacts *Contacts
}

type Customer struct {
	CustomerID   string
	Person       *Person
	Company      *Company `bson:",omitempty"`
	Localization *Localization
	Custom       interface{} `bson:",omitempty"`
}

type AddressType string

const (
	AddressTypeDelivery AddressType = "delivery"
	AddressTypeBilling  AddressType = "billing"
)

type Address struct {
	Person        *Person
	Type          AddressType // e.g. Shipping or Billing Address
	Street        string
	StreetNumber  string
	ZIP           string
	City          string
	Country       string
	PostOfficeBox string `bson:",omitempty"`
	//Extra        *AddressExtra
	Custom interface{} `bson:",omitempty"`
}

type AddressExtra struct {
	PostOfficeBox string
	// e.g. Packstation etc.
}

type Localization struct {
	LanguageCode LanguageCode
	CountryCode  CountryCode
}

type LanguageCode string

const (
	LanguageCodeGermany      LanguageCode = "DE"
	LanguageCodeSwistzerland LanguageCode = "CH"
)

type CountryCode string

const (
	CountryCodeGermany      CountryCode = "DE"
	CountryCodeSwistzerland CountryCode = "CH"
)
