package address

import (
	"strings"
)

// Person is a field Customer and of Address
// Only Customer->Person has Contacts
type Person struct {
	FirstName       string
	MiddleName      string
	LastName        string
	Title           TitleType
	Salutation      SalutationType
	Birthday        string
	Contacts        map[string]*Contact    // key must be contactID
	DefaultContacts map[ContactType]string // reference by contactID
}

func (person *Person) TrimSpace() {
	if person == nil {
		return
	}
	person.FirstName = strings.TrimSpace(person.FirstName)
	person.MiddleName = strings.TrimSpace(person.MiddleName)
	person.LastName = strings.TrimSpace(person.LastName)
	person.Title = TitleType(strings.TrimSpace(string(person.Title)))
	person.Salutation = SalutationType(strings.TrimSpace(string(person.Salutation)))
	person.Birthday = strings.TrimSpace(person.Birthday)
}
