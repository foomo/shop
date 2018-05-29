package address

import (
	"github.com/foomo/shop/unique"
	"gopkg.in/mgo.v2/bson"
)

//ContactsDeprecated is an old structure how contact data was stored in the past => it's deprecated
type ContactsDeprecated struct {
	PhoneLandLine string
	PhoneMobile   string
	Email         string
	Skype         string
	Primary       ContactType
}

// SetBSON to unmarshal BSON in a Person struct while migrating old objects into new person struct
func (p *Person) SetBSON(raw bson.Raw) error {
	return bsonDecodeNewPersonStruct(p, raw)
}

func bsonDecodeOldPersonStruct(p *Person, raw bson.Raw) error {
	// expected struct (old)
	decodedOld := new(struct {
		FirstName  string
		MiddleName string
		LastName   string
		Title      TitleType
		Salutation SalutationType
		Birthday   string
		Contacts   *ContactsDeprecated
	})

	// unmarshal
	bsonErr := raw.Unmarshal(decodedOld)

	// error
	if bsonErr != nil {
		return bsonErr
	}

	// map values
	p.FirstName = decodedOld.FirstName
	p.MiddleName = decodedOld.MiddleName
	p.LastName = decodedOld.LastName
	p.Title = decodedOld.Title
	p.Salutation = decodedOld.Salutation
	p.Birthday = decodedOld.Birthday
	p.Contacts = map[string]*Contact{}
	p.DefaultContacts = map[ContactType]string{}

	appendContact := func(p *Person, contactValue string, contactType ContactType) {
		if contactValue != "" {
			id := unique.GetNewIDShortID()
			contact := &Contact{
				ID:    id,
				Type:  contactType,
				Value: contactValue,
			}
			p.Contacts[contact.ID] = contact
			p.DefaultContacts[contactType] = contact.ID
		}
	}

	appendContact(p, decodedOld.Contacts.Email, ContactTypeEmail)
	appendContact(p, decodedOld.Contacts.PhoneLandLine, ContactTypePhoneLandline)
	appendContact(p, decodedOld.Contacts.PhoneMobile, ContactTypePhoneMobile)
	appendContact(p, decodedOld.Contacts.Skype, ContactTypeSkype)

	// no error
	return nil
}

func bsonDecodeNewPersonStruct(p *Person, raw bson.Raw) error {
	// expected struct
	decoded := new(struct {
		FirstName       string
		MiddleName      string
		LastName        string
		Title           TitleType
		Salutation      SalutationType
		Birthday        string
		Contacts        map[string]*Contact
		DefaultContacts map[ContactType]string
	})

	// unmarshall
	bsonErr := raw.Unmarshal(decoded)

	// error
	if bsonErr != nil {
		return bsonErr
	}

	// no contacts decoded, try to decode old struct instead
	if decoded.Contacts == nil {
		return bsonDecodeOldPersonStruct(p, raw)
	}

	// map values
	p.FirstName = decoded.FirstName
	p.MiddleName = decoded.MiddleName
	p.LastName = decoded.LastName
	p.Title = decoded.Title
	p.Salutation = decoded.Salutation
	p.Birthday = decoded.Birthday
	p.Contacts = decoded.Contacts
	p.DefaultContacts = decoded.DefaultContacts

	// no error
	return nil
}
