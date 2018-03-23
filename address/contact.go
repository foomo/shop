package address

type Contact struct {
	ID        string
	Type      ContactType
	Value     string
	IsDefault bool
}

// IsPhone if type is in (ContactTypePhone, ContactTypePhoneMobile, ContactTypePhoneLandline)
func (c *Contact) IsPhone() bool {
	return c.Type == ContactTypePhone || c.Type == ContactTypePhoneMobile || c.Type == ContactTypePhoneLandline
}

// GetPhone returns a phone or nil if none is stored yet
// tries to find types in the given order: ContactTypePhone first, ContactTypePhoneMobile second, ContactTypePhoneLandline last
// if one of these contact types is marked as default it will be returned immediately
func (p *Person) GetPhone() *Contact {

	// fallback
	var fallback *Contact
	setFallback := func(contact *Contact) {
		if fallback == nil {
			fallback = contact
		}
	}

	// generic phone
	phone := p.GetContactByType(ContactTypePhone)
	if phone != nil && phone.IsDefault {
		return phone
	}
	setFallback(phone)

	// mobile
	mobile := p.GetContactByType(ContactTypePhoneMobile)
	if mobile != nil && mobile.IsDefault {
		return mobile
	}
	setFallback(mobile)

	// landline
	landline := p.GetContactByType(ContactTypePhoneLandline)
	if landline != nil && landline.IsDefault {
		return landline
	}
	setFallback(landline)

	return fallback
}

// GetContactByType will search for given contact type
// if possible the default entry for that type will be returned
// if no default exists at least any of that type will be returned
// if none of that type exists it will be nil
func (p *Person) GetContactByType(contactType ContactType) *Contact {
	var match *Contact
	for _, c := range p.Contacts {
		if c.Type == contactType {
			if c.IsDefault {
				return c
			}
			if match == nil {
				match = c
			}
		}
	}
	return match
}
