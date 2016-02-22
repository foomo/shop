package customer

type Customer struct {
	FirstName   string
	LastName    string
	Title       string
	Salutation  string
	Email       string
	PhoneNumber string
	Address     Address
	Custom      interface{}
}

type Address struct {
	Description  string // e.g. Shipping or Billing Address
	Street       string
	StreetNumber string
	ZIP          string
	City         string
	Country      string
	Custom       interface{}
}
