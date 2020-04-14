package customer

import (
	"errors"

	"github.com/foomo/shop/shop_error"
)

// ErrCustomerNotFound if customer not found
var ErrCustomerNotFound = errors.New(shop_error.ErrorNotInDatabase)
