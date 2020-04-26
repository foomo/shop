package address

import (
	"math/rand"
	"testing"
	"time"
)

type fields struct {
	Id           string
	ExternalID   string
	Person       *Person
	Type         AddressType
	Street       string
	StreetNumber string
	ZIP          string
	City         string
	Country      string
	CountryCode  string
	Company      string
}

func TestAddress_Validate(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"valid address",
			validAddressFields(),
			false,
		},
		{
			"invalid person phone",
			invalidPersonContactsPhoneGt250(),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := &Address{
				Id:           tt.fields.Id,
				ExternalID:   tt.fields.ExternalID,
				Person:       tt.fields.Person,
				Type:         tt.fields.Type,
				Street:       tt.fields.Street,
				StreetNumber: tt.fields.StreetNumber,
				ZIP:          tt.fields.ZIP,
				City:         tt.fields.City,
				Country:      tt.fields.Country,
				CountryCode:  tt.fields.CountryCode,
				Company:      tt.fields.Company,
			}
			if err := address.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func invalidPersonContactsPhoneGt250() fields {
	f := validAddressFields()
	f.Person = validPerson()
	f.Person.SetPhone(randSeq(255))
	return f
}

func validAddressFields() fields {
	return fields{
		Id:         randSeq(10),
		ExternalID: randSeq(10),
		Person: &Person{
			FirstName:  randSeq(10),
			LastName:   randSeq(10),
			Salutation: SalutationTypeMr,
			Contacts: map[string]*Contact{
				randSeq(10): {
					Type:  ContactTypePhone,
					Value: randSeq(10),
				},
			},
		},
		Type:         AddressDefaultBilling,
		Street:       randSeq(10),
		StreetNumber: randSeq(10),
		ZIP:          randSeq(4),
		City:         randSeq(10),
		Country:      randSeq(10),
		CountryCode:  randSeq(10),
		Company:      randSeq(10),
	}
}

func validPerson() *Person {
	return &Person{
		FirstName:  randSeq(10),
		LastName:   randSeq(10),
		Salutation: SalutationTypeMr,
		Contacts: map[string]*Contact{
			randSeq(10): {
				Type:  ContactTypePhone,
				Value: randSeq(10),
			},
		},
	}
}
