package customer

type base struct {
	Type      string
	FirstName string
	LastName  string
	Title     string
	Email     string
}

type Customer struct {
	base
	Custom interface{}
}

type Address struct {
	base
}
