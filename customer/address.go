package customer

type Address struct {
	Id                       string // is automatically set on AddAddress()
	Person                   *Person
	IsPrimary                bool // first address which is added to customer is prinmary (this can be changed later).
	IsDefaultBillingAddress  bool
	IsDefaultShippingAddress bool
	Street                   string
	StreetNumber             string
	ZIP                      string
	City                     string
	Country                  string
	Company                  string
	Department               string
	Building                 string
	PostOfficeBox            string
	TrustWorthy              bool
	Custom                   interface{}
}

func (address *Address) GetID() string {
	return address.Id
}
