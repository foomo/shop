package address

import (
	"strings"

	"github.com/foomo/shop/unique"
)

// IsMail if type is ContactTypeEmail
func (c *Contact) IsMail() bool {
	return c.Type == ContactTypeEmail
}

// CreateMailContact will return a mail contact for given mail address, a unique ID is generated automatically
func CreateMailContact(mail string) *Contact {
	return &Contact{
		ID:    unique.GetNewIDShortID(),
		Type:  ContactTypePhone,
		Value: strings.ToLower(mail),
	}
}

// SetMail will set or replace the default mail. If no default mail is present it will replace first mail contact. Otherwise a new default mail contact is created.
func (p *Person) SetMail(mail string) *Contact {
	contact := p.GetContactByType(ContactTypeEmail)
	if contact == nil {
		contact = CreateMailContact(mail)
		p.Contacts[contact.ID] = contact
		p.DefaultContacts[ContactTypeEmail] = contact.ID
	}
	contact.Value = strings.ToLower(mail)
	return contact
}

// GetMailAddress will return the email address string or an empty string if not yet set
func (p *Person) GetMailAddress() string {
	mail := p.GetMailContact()
	if mail != nil {
		return mail.Value
	}
	return ""
}

// GetMailContact will return the default contact object for contactType mail or nil of not yet set
func (p *Person) GetMailContact() *Contact {
	return p.GetContactByType(ContactTypeEmail)
}
