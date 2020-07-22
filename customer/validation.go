package customer

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// IsCustomerComplete returns an error if not all mandatory data is set
func (customer *Customer) IsComplete() error {
	addr, err := customer.GetDefaultBillingAddress()
	if err != nil {
		return fmt.Errorf("could not load default billing address: %q", err)
	}

	var mErr *multierror.Error
	if e := customer.GetEmail(); !strings.ContainsRune(e, '@') {
		mErr = multierror.Append(mErr, fmt.Errorf("invalid email address: %q", e))
	}

	if err := addr.IsComplete(); err != nil {
		mErr = multierror.Append(mErr, fmt.Errorf("address is not complete: %q", err))
	}

	person := customer.GetPerson()
	if err := person.IsComplete(); err != nil {
		mErr = multierror.Append(mErr, fmt.Errorf("person is not complete: %q", err))
	}

	// Birthday is not part of regular person.IsComplete() check
	if person != nil {
		if len(person.Birthday) != 10 {
			mErr = multierror.Append(mErr, fmt.Errorf("person birthday not valid"))
		}
	}

	return mErr.ErrorOrNil()
}
