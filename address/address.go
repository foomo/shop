package address

import (
	"errors"
	"fmt"
	"gopkg.in/validator.v2"
	"reflect"
	"strconv"
)

type AddressType string
type ContactType string
type SalutationType string
type TitleType string

const (
	AddressDefaultBilling  AddressType = "addresDefaultBilling"
	AddressDefaultShipping AddressType = "addressDefaultShipping"
	AddressOther           AddressType = "addressOther"

	ContactTypeEmail         ContactType = "email"
	ContactTypePhone         ContactType = "phone"
	ContactTypePhoneMobile   ContactType = "mobile"
	ContactTypePhoneLandline ContactType = "landline"
	ContactTypeFax           ContactType = "fax"
	ContactTypeSkype         ContactType = "skype"
	ContactTypeSlack         ContactType = "slack"
	ContactTypeTwitter       ContactType = "twitter"
	ContactTypeFacebook      ContactType = "facebook"

	SalutationTypeMr       SalutationType = "male"   //"Mr"
	SalutationTypeMrs      SalutationType = "female" //"Mrs"
	SalutationTypeMrAndMrs SalutationType = "MrAndMrs"
	SalutationTypeCompany  SalutationType = "Company" // TODO: find better wording
	SalutationTypeFamily   SalutationType = "Family"  // TODO: find better wording

	TitleTypeNone   TitleType = ""
	TitleTypeDr     TitleType = "Dr"
	TitleTypeProf   TitleType = "Prof."
	TitleTypeProfDr TitleType = "Prof. Dr."
	TitleTypePriest TitleType = "Priest" // TODO: find better wording
)

type Address struct {
	Id            string      `validate:"nonzero"` // is automatically set on AddAddress()
	ExternalID    string      `validate:"nonzero"`
	Person        *Person     `validate:"nonzero"`
	Type          AddressType `validate:"nonzero"`
	Street        string      `validate:"nonzero, min=2, max=189"`
	StreetNumber  string      `validate:"nonzero, max=60"`
	ZIP           string      `validate:"nonzero, min=4,max=5"`
	City          string      `validate:"nonzero, min=3,max=60"`
	Country       string      `validate:"nonzero"`
	CountryCode   string      `validate:"nonzero"`
	Company       string      `validate:"max=250"`
	Department    string
	Building      string
	PostOfficeBox string
	Score         *Score
	Custom        interface{}
}

// Person is a field Customer and of Address
// Only Customer->Person has Contacts
type Person struct {
	FirstName       string `validate:"nonzero, max=60"`
	MiddleName      string
	LastName        string `validate:"nonzero, max=60"`
	Title           TitleType
	Salutation      SalutationType `validate:"nonzero"`
	Birthday        string
	Contacts        map[string]*Contact    `validate:"vpMax=250"` // key must be contactID
	DefaultContacts map[ContactType]string // reference by contactID
}

func (address *Address) GetID() string {
	return address.Id
}

type Score struct {
	Trusted          bool
	TrustedString    string
	DateOfTrustCheck string
	ScoringFailed    bool
}

func (address *Address) HasScore() bool {
	if address.Score == nil || address.Score.DateOfTrustCheck == "" {
		return false
	}
	return true
}

// Equals checks if this is BASICALLY the same address. (Type may be different).
func (address *Address) Equals(otherAddress *Address) bool {
	equal := true

	equal = equal && address.Person.FirstName == otherAddress.Person.FirstName
	equal = equal && address.Person.LastName == otherAddress.Person.LastName
	equal = equal && address.Street == otherAddress.Street
	equal = equal && address.StreetNumber == otherAddress.StreetNumber
	equal = equal && address.ZIP == otherAddress.ZIP
	equal = equal && address.City == otherAddress.City
	equal = equal && address.Country == otherAddress.Country

	return equal
}

// validatePhoneMax validates Contacts. If there os contact with type Phone present, it checks for maximum characters length
func validatePhoneMax(v interface{}, param string) error {
	max, _ := strconv.Atoi(param)
	st := reflect.ValueOf(v)
	keys := st.MapKeys()
	for _, key := range keys {
		v := st.MapIndex(key)
		ttype := v.Interface().(*Contact).Type
		value := v.Interface().(*Contact).Value
		if ttype == "" {
			return errors.New("type must be set")
		}
		if ttype == ContactTypePhone {
			if len(value) > max {
				return fmt.Errorf("phone cannot be longer then %d characters", max)
			}
		}
	}
	return nil
}

// IsValid checks if address contents are valid based on validation specified
func (address *Address) Validate() error {
	validator.SetValidationFunc("vpMax", validatePhoneMax)
	return validator.Validate(address)
}
