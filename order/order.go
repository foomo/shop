// Package handles order in a shop
// - carts are incomplete orders
package order

import (
	"github.com/foomo/shop/customer"
	"github.com/foomo/shop/payment"
	"gopkg.in/mgo.v2/bson"
)

type Position struct {
}

type Order struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Positions []*Position
	Customer  *customer.Customer
	Addresses []*customer.Address
	Payments  []*payment.Payment
	Custom    interface{}
}

func NewOrder(customOrder interface{}) *Order {
	return &Order{
		Custom: customOrder,
	}
}

func (o *Order) GetCustomer(customCustomer interface{}) (c *customer.Customer, err error) {
	c = &customer.Customer{
		Custom: customCustomer,
	}
	// do mongo magic to load customCustomer
	return

}
