package address

type AddressType string
type ContactType string
type SalutationType string
type TitleType string

const (
	AddressDefaultBilling    AddressType    = "addresDefaultBilling"
	AddressDefaultShipping   AddressType    = "addressDefaultShipping"
	AddressOther             AddressType    = "addressOther"
	ContactTypePhoneLandline ContactType    = "landline"
	ContactTypePhoneMobile   ContactType    = "mobile"
	ContactTypeEmail         ContactType    = "email"
	ContactTypeSkype         ContactType    = "skype"
	ContactTypeFax           ContactType    = "fax"
	SalutationTypeMr         SalutationType = "male"   //"Mr"
	SalutationTypeMrs        SalutationType = "female" //"Mrs"
	SalutationTypeMrAndMrs   SalutationType = "MrAndMrs"
	SalutationTypeCompany    SalutationType = "Company" // TODO: find better wording
	SalutationTypeFamily     SalutationType = "Family"  // TODO: find better wording
	TitleTypeDr              TitleType      = "Dr"
	TitleTypeProf            TitleType      = "Prof."
	TitleTypeProfDr          TitleType      = "Prof. Dr."
	TitleTypePriest          TitleType      = "Priest" // TODO: find better wording
)

type Address struct {
	Id     string // is automatically set on AddAddress()
	Person *Person
	// IsDefaultBillingAddress  bool
	// IsDefaultShippingAddress bool
	Type          AddressType
	Street        string
	StreetNumber  string
	ZIP           string
	City          string
	Country       string
	Company       string
	Department    string
	Building      string
	PostOfficeBox string
	Score         *Score
	Custom        interface{}
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
type Contacts struct {
	PhoneLandLine string
	PhoneMobile   string
	Email         string
	Skype         string
	Primary       ContactType
}

func (address *Address) GetID() string {
	return address.Id
}

type Score struct {
	Trusted          bool
	TrustedString    string
	DateOfTrustCheck string
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

// GetPrimaryContact returns primary contact as string
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
		// case ContactTypeFax: // 2016 anyone??
		// 	return string(ContactTypeFax) + ": " + c.Fax
	}
	return "No primary contact available!"
}
