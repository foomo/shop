package address

type Contact struct {
	ID         string
	ExternalID string
	Type       ContactType
	Value      string
}

// GetContactByType will search for given contact type
// if possible the default entry for that type will be returned
// if no default exists at least any of that type will be returned
// if none of that type exists it will be nil
func (p *Person) GetContactByType(contactType ContactType) *Contact {

	// try to load default first
	defaultContact := p.GetDefaultContactByType(contactType)
	if defaultContact != nil {
		return defaultContact
	}

	// no default ... try to find a fallback of same type
	for _, c := range p.Contacts {
		if c.Type == contactType {
			return c
		}
	}

	// we failed ... contact type is not present
	return nil
}

func (p *Person) GetDefaultContactByType(contactType ContactType) *Contact {
	// try to return default
	if id, ok := p.DefaultContacts[contactType]; ok {
		if c, ok := p.Contacts[id]; ok {
			return c
		}
	}

	return nil
}
