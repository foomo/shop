package customer

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
	Company                  string
	Department               string
	Building                 string
	PostOfficeBox            string
	Custom                   interface{}
}

func (address *Address) GetID() string {
	return address.Id
}
