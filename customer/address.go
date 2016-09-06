package customer

type AddressType string

const (
	AddressDefaultBilling  AddressType = "addresDefaultBilling"
	AddressDefaultShipping AddressType = "addressDefaultShipping"
	AddressOther           AddressType = "addressOther"
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
