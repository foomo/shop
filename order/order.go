// Package order handles order in a shop
// - carts are incomplete orders
//
package order

import (
	"time"

	"github.com/foomo/shop/customer"
	"github.com/foomo/shop/payment"
	"gopkg.in/mgo.v2/bson"
)

// Event can be triggered and listended to in order to deal with changes of\
// orders
type Event struct {
	Type      string
	Comment   string
	Timestamp time.Time
	Custom    interface{}
}

// Order of item
// create revisions
type Order struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Status    string
	History   []*Event
	Positions []*Position
	Customer  *customer.Customer
	Addresses []*customer.Address
	Payments  []*payment.Payment
	Custom    interface{}
}

// OrderCustom custom object provider
type OrderCustomProvider interface {
	NewOrderCustom() interface{}
	NewPositionCustom() interface{}
	NewAddressCustom() interface{}
}

// NewOrder
func NewOrder(customOrder interface{}) *Order {
	return &Order{
		Custom: customOrder,
	}
}

// GetCustomer
func (o *Order) GetCustomer(customCustomer interface{}) (c *customer.Customer, err error) {
	c = &customer.Customer{
		Custom: customCustomer,
	}
	// do mongo magic to load customCustomer
	return
}

func (o *Order) SaveRevision(revisionInfo interface{}) error {
	return nil
}
