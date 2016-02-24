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
	PhoneLandline string
	PhoneMobile   string
	Email         string
	Skype         string
	Fax           string
	Primary       ContactType
}

//
// func NewPerson(title, salutation, firstName, lastName) *Person {
// 	return &Person{
// 		Title:      title,
// 		Salutation: salutation,
// 		FirstName:  firstName,
// 		LastName:   lastName,
// 		Contacts:   make(map[ContactType]*Contact),
// 	}
// }

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
	Custom       interface{} `bson:"omitempty"`
}

type AddressType string

const (
	AddressTypeDelivery AddressType = "delivery"
	AddressTypeBilling  AddressType = "billing"
)

type Address struct {
	Person       *Person
	Type         AddressType // e.g. Shipping or Billing Address
	Street       string
	StreetNumber string
	ZIP          string
	City         string
	Country      string
	Extra        *AddressExtra `bson:",omitempty"`
	Custom       interface{}
}

type AddressExtra struct {
	PostOfficeBox string
	// e.g. Packstation etc.
}

type Localization struct {
	LanguageCode string
	CountryCode  string
}
