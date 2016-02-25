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
	"github.com/foomo/shop/shipping"
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

type OrderPriceInfo struct {
	SumNet        float64
	RebatesNet    float64
	VouchersNet   float64
	ShippingNet   float64
	SumFinalNet   float64
	Taxes         float64
	SumFinalGross float64
}

// Order of item
// create revisions
type Order struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	OrderID   string
	OrderType OrderType
	Timestamp time.Time
	Status    OrderStatus
	History   []*Event
	Positions []*Position
	Customer  *customer.Customer
	Addresses []*customer.Address
	Payment   *payment.Payment
	PriceInfo *OrderPriceInfo
	Shipping  *shipping.ShippingProperties
	Custom    interface{} `bson:",omitempty"`
	Queue     *struct {
		Name           string
		RetryAfter     time.Duration
		LastProcessing time.Time

		//BulkID string
	}
}

type OrderType string

const (
	OrderTypeOrder  OrderType = "order"
	OrderTypeReturn OrderType = "return"
)

type OrderStatus string

const (
	OrderStatusCreated  OrderStatus = "created"
	OrderStatusPocessed OrderStatus = "processed"
	OrderStatusShipped  OrderStatus = "shipped"
)

// OrderCustomProvider custom object provider
type OrderCustomProvider interface {
	NewOrderCustom() interface{}
	NewPositionCustom() interface{}
	NewAddressCustom() interface{}
	NewCustomerCustom() interface{}
	Fields() *bson.M
}

// NewOrder
func NewOrder(customOrder interface{}) *Order {
	return &Order{
		Timestamp: time.Now(),
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
