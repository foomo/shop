// Package order handles order in a shop
// - carts are incomplete orders
//
package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	foomo_shop_configuration "github.com/foomo/shop/configuration"
	"github.com/foomo/shop/customer"
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/payment"
	"github.com/foomo/shop/shipping"
	"gopkg.in/mgo.v2/bson"
)

// Event can be triggered and listended to in order to deal with changes of\
// orders

type ActionOrder string

const (
	ActionStatusUpdateHead       ActionOrder = "actionStatusUpdateHead"
	ActionStatusUpdatePosition   ActionOrder = "actionStatusUpdatePosition"
	ActionNoATPResponseForItemID ActionOrder = "actionNoATPResponseForItemID"
	ActionValidateStatusHead     ActionOrder = "actionValidateStatusHead"
	ActionValidateStatusPosition ActionOrder = "actionValidateStatusPosition"
	ActionAddPosition            ActionOrder = "actionAddPosition"
	ActionRemovePosition         ActionOrder = "actionRemovePosition"
	ActionChangeQuantityPosition ActionOrder = "actionChangeQuantityPosition"
	ActionCreateCustomOrder      ActionOrder = "actionCreateCustomOrder"
)

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
	History   event_log.EventHistory
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
	OrderStatusInvalid  OrderStatus = "invalid status"
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
func NewOrder() *Order {
	return &Order{
		Timestamp: time.Now(),
		History:   event_log.EventHistory{},
		Positions: []*Position{},
		Customer:  &customer.Customer{},
		Addresses: []*customer.Address{},
		Payment:   &payment.Payment{},
		PriceInfo: &OrderPriceInfo{},
		Shipping:  &shipping.ShippingProperties{},
	}
}

func (o *Order) SaveOrderEvent(action ActionOrder, err error, positionItemNumber string) {
	o.SaveOrderEventDetailed(action, err, "", "")
}

func (o *Order) SaveOrderEventDetailed(action ActionOrder, err error, positionItemNumber string, comment string) {
	event_log.Debug("Action", string(action), "OrderID", o.OrderID)
	event := event_log.NewEvent()
	if err != nil {
		event.Type = event_log.EventTypeError
	} else {
		event.Type = event_log.EventTypeSuccess
	}
	event.Action = string(action)
	event.OrderID = o.OrderID
	event.PositionItemID = positionItemNumber
	event.Comment = comment
	if err != nil {
		event.Error = err.Error()
	}
	o.History = append(o.History, event)
	GetPersistor(foomo_shop_configuration.MONGO_URL, foomo_shop_configuration.MONGO_COLLECTION_ORDERS).UpsertOrder(o) // Error is ignored because it gets already logged in UpsertOrder()

	jsonBytes, _ := json.MarshalIndent(event, "", "	")
	event_log.Debug("Saved Shop Event! ", string(jsonBytes))
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

	existingPos := o.GetPositionByItemId(pos.ItemID)
	if existingPos != nil {
		err := errors.New("position already exists use SetPositionQuantity or GetPositionById to manipulate it")
		o.SaveOrderEventDetailed(ActionAddPosition, err, pos.ItemID, "")
		return err
	}
	o.Positions = append(o.Positions, pos)
	if _, err := GetPersistor(foomo_shop_configuration.MONGO_URL, foomo_shop_configuration.MONGO_COLLECTION_ORDERS).UpsertOrder(o); err != nil {
		o.SaveOrderEventDetailed(ActionAddPosition, err, pos.ItemID, "Error while adding position to order. Could not upsert order")
		return err
	}
	return nil
}

func (o *Order) SetPositionQuantity(itemID string, quantity float64) error {
	pos := o.GetPositionByItemId(itemID)
	if pos == nil {
		err := fmt.Errorf("position with %q not found in order", itemID)
		o.SaveOrderEventDetailed(ActionChangeQuantityPosition, err, pos.ItemID, "Could not set quantity to "+fmt.Sprint(quantity))
		return err
	}
	pos.Quantity = quantity
	o.SaveOrderEventDetailed(ActionChangeQuantityPosition, nil, pos.ItemID, "Set quantity to "+fmt.Sprint(quantity))
	// remove position if quantity is zero
	if pos.Quantity == 0.0 {
		for index := range o.Positions {
			if pos.ItemID == itemID {
				o.Positions = append(o.Positions[:index], o.Positions[index+1:]...)
				return nil
			}
		}
	}
	if _, err := GetPersistor(foomo_shop_configuration.MONGO_URL, foomo_shop_configuration.MONGO_COLLECTION_ORDERS).UpsertOrder(o); err != nil {
		o.SaveOrderEventDetailed(ActionChangeQuantityPosition, err, pos.ItemID, "Error while updating position quantity. Could not upsert order")
		return err
	}
	return nil
}

func (o *Order) GetPositionByItemId(itemID string) *Position {
	for _, pos := range o.Positions {
		if pos.ItemID == itemID {
			return pos
		}
	}
	return nil
}
