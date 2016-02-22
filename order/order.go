// Package order handles order in a shop
// - carts are incomplete orders
//
package order

import (
	"errors"
	"fmt"
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
	Timestamp time.Time
	Status    string
	Queue     *struct {
		Name           string
		RetryAfter     time.Duration
		LastProcessing time.Time

		//BulkID string
	}
	History   []*Event
	Positions []*Position
	Customer  *customer.Customer
	//Addresses []*customer.Address
	Payments []*payment.Payment
	Custom   interface{}
}

// OrderCustomProvider custom object provider
type OrderCustomProvider interface {
	NewOrderCustom() interface{}
	NewPositionCustom() interface{}
	NewAddressCustom() interface{}
	Fields() *bson.M
}

// NewOrder
func NewOrder(customOrder interface{}) *Order {
	return &Order{
		Timestamp: time.Now(),
		Custom:    customOrder,
	}
}

func (o *Order) AddEventToHistory(e *Event) {
	o.History = append(o.History, e)
}

// GetCustomer
func (o *Order) GetCustomer(customCustomer interface{}) (c *customer.Customer, err error) {
	c = &customer.Customer{
		Custom: customCustomer,
	}
	// do mongo magic to load customCustomer
	return
}

func (o *Order) AddPosition(pos *Position) error {
	existingPos := o.GetPositionById(pos.ID)
	if existingPos != nil {
		return errors.New("position already exists use SetPositionQuantity or GetPositionById to manipulate it")
	}
	o.Positions = append(o.Positions, pos)
	return nil
}

func (o *Order) SetPositionQuantity(id string, quantity float64) error {
	pos := o.GetPositionById(id)
	if pos == nil {
		return fmt.Errorf("position with %q not found in order", id)
	}
	pos.Quantity = quantity
	if pos.Quantity == 0.0 {
		positions := []*Position{}
		for _, pos := range o.Positions {
			if pos.ID == id {
				continue
			}
			positions = append(positions, pos)
		}
		o.Positions = positions

	}
	return nil
}

func (o *Order) GetPositionById(id string) *Position {
	for _, pos := range o.Positions {
		if pos.ID == id {
			return pos
		}
	}
	return nil
}

func (o *Order) SaveRevision(revisionInfo interface{}) error {
	return nil
}

//func (o *Order)Queue
