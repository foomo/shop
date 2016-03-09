package payment

import "time"

type CreditCard struct {
	FirstName         string
	LastName          string
	CardNumber        string
	ExpirationDate    string
	AuthorizationCode string
}

type Debit struct {
	FirstName       string
	LastName        string
	IBAN            string
	BIC             string
	NameOfInstitute string
}

type Voucher struct {
	Code           string
	Value          float64
	ExpirationDate time.Time
}
