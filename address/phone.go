package address

import "github.com/foomo/shop/unique"

// IsPhone if type is in (ContactTypePhone, ContactTypePhoneMobile, ContactTypePhoneLandline)
func (c *Contact) IsPhone() bool {
	return c.Type == ContactTypePhone || c.Type == ContactTypePhoneMobile || c.Type == ContactTypePhoneLandline
}

// CreatePhoneContact will return a phone contact for given phone number, a unique ID is generated automatically
func CreatePhoneContact(phone string) *Contact {
	return &Contact{
		ID:    unique.GetNewIDShortID(),
		Type:  ContactTypePhone,
		Value: phone,
	}
}

// SetPhone will set or replace the default phone. If no default phone is present it will replace first phone contact. Otherwise a new default phone contact is created.
func (p *Person) SetPhone(phone string) *Contact {
	contact := p.GetContactByType(ContactTypePhone)
	if contact == nil {
		if p.Contacts == nil {
			p.Contacts = map[string]*Contact{}
			p.DefaultContacts = map[ContactType]string{}
		}
		contact = CreatePhoneContact(phone)
		p.Contacts[contact.ID] = contact
		p.DefaultContacts[ContactTypePhone] = contact.ID
	}
	contact.Value = phone
	return contact
}

// GetPhoneNumber will return the default phone number or an empty string if not yet set
func (p *Person) GetPhoneNumber() string {
	phone := p.GetPhoneContact()
	if phone != nil {
		return phone.Value
	}
	return ""
}

// GetPhoneContact returns a phone contact or nil if none is stored yet
// tries to find types in the given order: ContactTypePhone first, ContactTypePhoneMobile second, ContactTypePhoneLandline last
// if one of these contact types is marked as default it will be returned immediately
func (p *Person) GetPhoneContact() *Contact {

	// try to load default contact first for phone or mobile or landline (in that exact order)
	types := []ContactType{ContactTypePhone, ContactTypePhoneMobile, ContactTypePhoneLandline}
	for _, contactType := range types {
		defaultContact := p.GetDefaultContactByType(contactType)
		if defaultContact != nil {
			return defaultContact
		}
	}

	// try to load any phone contact in given order: phone, mobile, landline
	for _, contactType := range types {
		contact := p.GetContactByType(contactType)
		if contact != nil {
			return contact
		}
	}

	// no phone contact
	return nil
}
