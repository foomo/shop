package address

import (
	"testing"

	"github.com/foomo/shop/unique"

	"github.com/stretchr/testify/assert"

	"gopkg.in/mgo.v2/bson"
)

func TestPersonContactsMigration(t *testing.T) {

	// marshall empty person

	var emptyPerson Person

	emptyPersonMarshal, emptyPersonMarshalErr := bson.Marshal(emptyPerson)
	assert.NoError(t, emptyPersonMarshalErr, "marshal emptyPerson failed")

	emptyPersonUnmarshal := &Person{}
	emptyPersonUnmarshalErr := bson.Unmarshal(emptyPersonMarshal, emptyPersonUnmarshal)
	assert.NoError(t, emptyPersonUnmarshalErr, "unmarshal emptyPerson failed")

	// marshall sepp

	sepp := &Person{}
	sepp.FirstName = "Sepp"

	seppMarshal, seppMarshalErr := bson.Marshal(sepp)
	assert.NoError(t, seppMarshalErr, "marshal sepp failed")

	seppUnmarshal := &Person{}
	seppUnmarshalErr := bson.Unmarshal(seppMarshal, seppUnmarshal)
	assert.NoError(t, seppUnmarshalErr, "unmarshal sepp failed")

	// marshall max

	p := &Person{
		Salutation: SalutationTypeMr,
		Title:      TitleTypeDr,
		FirstName:  "Max",
		LastName:   "Mustermann",
		Birthday:   "1973-12-22",
		Contacts: []*Contact{
			&Contact{
				ID:        unique.GetNewIDShortID(),
				IsDefault: true,
				Type:      ContactTypeEmail,
				Value:     "foo@example.com",
			},
			&Contact{
				ID:        unique.GetNewIDShortID(),
				IsDefault: true,
				Type:      ContactTypePhone,
				Value:     "+49 1234 56 78 990",
			},
		},
	}

	m, marshalErr := bson.Marshal(p)
	assert.NoError(t, marshalErr, "marshal person failed")

	u := &Person{}
	unmarshalErr := bson.Unmarshal(m, u)
	assert.NoError(t, unmarshalErr, "unmarshal person failed")

	assert.Equal(t, p.Salutation, u.Salutation)
	assert.Equal(t, p.Title, u.Title)
	assert.Equal(t, p.FirstName, u.FirstName)
	assert.Equal(t, p.LastName, u.LastName)
	assert.Equal(t, p.Birthday, u.Birthday)

	assert.Len(t, p.Contacts, 2, "expected two contact types")
	assert.Len(t, u.Contacts, 2, "expected two contact types")

	type OldPersonStruct struct {
		FirstName  string
		MiddleName string
		LastName   string
		Title      TitleType
		Salutation SalutationType
		Birthday   string
		Contacts   *ContactsDeprecated
	}

	old := &OldPersonStruct{
		Salutation: p.Salutation,
		Title:      p.Title,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Birthday:   p.Birthday,
		Contacts: &ContactsDeprecated{
			Email:       "foo@example.com",
			PhoneMobile: "+49 1234 56 78 990",
		},
	}

	m, marshalErr = bson.Marshal(old)
	assert.NoError(t, marshalErr, "marshal person failed")

	u = &Person{}
	unmarshalErr = bson.Unmarshal(m, u)
	assert.NoError(t, unmarshalErr, "unmarshal person failed")

	assert.NotNil(t, u.Contacts, "contacts must not be empty")
	assert.Len(t, u.Contacts, 2, "expected two contact types")
}
