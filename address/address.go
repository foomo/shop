package address

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

	TitleTypeDr     TitleType = "Dr"
	TitleTypeProf   TitleType = "Prof."
	TitleTypeProfDr TitleType = "Prof. Dr."
	TitleTypePriest TitleType = "Priest" // TODO: find better wording
)

type Address struct {
	Id            string // is automatically set on AddAddress()
	Person        *Person
	Type          AddressType
	Street        string
	StreetNumber  string
	ZIP           string
	City          string
	Country       string
	CountryCode   string
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
	Contacts   []*Contact
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
