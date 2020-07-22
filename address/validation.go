package address

import (
	"errors"

	"github.com/hashicorp/go-multierror"
)

func (addr *Address) IsComplete() error {
	addr.TrimSpace()
	// Return error if required field is missing
	var mErr *multierror.Error

	if addr == nil {
		mErr = multierror.Append(mErr, errors.New("address is nil"))
		return mErr.ErrorOrNil()
	}

	errPerson := addr.Person.IsComplete()
	if errPerson != nil {
		mErr = multierror.Append(mErr, errPerson)
	}

	if addr.Street == "" {
		mErr = multierror.Append(mErr, errors.New("address street is empty"))
	}
	if addr.StreetNumber == "" {
		mErr = multierror.Append(mErr, errors.New("address street number is empty"))
	}
	if len(addr.ZIP) < 4 {
		mErr = multierror.Append(mErr, errors.New("address zip is not valid"))
	}
	if addr.City == "" {
		mErr = multierror.Append(mErr, errors.New("address city is empty"))
	}
	if addr.Country == "" {
		mErr = multierror.Append(mErr, errors.New("address country is empty"))
	}
	if addr.CountryCode == "" {
		mErr = multierror.Append(mErr, errors.New("address country code is empty"))
	}

	return mErr.ErrorOrNil()
}

func (person *Person) IsComplete() error {
	var mErr *multierror.Error
	if person == nil {
		mErr = multierror.Append(mErr, errors.New("person is nil"))
		return mErr.ErrorOrNil()
	}
	person.TrimSpace()
	if person.Salutation == "" {
		mErr = multierror.Append(mErr, errors.New("person salutation is empty"))
	}
	if person.FirstName == "" {
		mErr = multierror.Append(mErr, errors.New("person firstname is empty"))
	}
	if person.LastName == "" {
		mErr = multierror.Append(mErr, errors.New("person lastname is empty"))
	}
	return mErr.ErrorOrNil()
}
