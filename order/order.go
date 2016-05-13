// Package order handles order in a shop
// - carts are incomplete orders
//
package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

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
	ActionValidation             ActionOrder = "actionValidation"
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
	//Addresses []*customer.Address
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
	OrderStatusInvalid     OrderStatus = "orderStatusInvalid"
	OrderStatusCreated     OrderStatus = "orderStatusCreated"
	OrderStatusPocessed    OrderStatus = "orderStatusProcessed"
	OrderStatusShipped     OrderStatus = "orderStatusShipped"
	OrderStatusReadyForATP OrderStatus = "orderStatusReadyForATP"
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
		//Addresses: []*customer.Address{},
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
	GetOrderPersistor().UpsertOrder(o) // Error is ignored because it gets already logged in UpsertOrder()

	jsonBytes, _ := json.MarshalIndent(event, "", "	")
	event_log.Debug("Saved Order Event! ", string(jsonBytes))
}

// Event will only be saved if is an error
func (o *Order) SaveOrderEventOnError(action ActionOrder, err error, positionItemNumber string, comment string) {
	if err == nil {
		return
	}
	o.SaveOrderEventDetailed(action, err, positionItemNumber, comment)
}

func (o *Order) SaveOrderEventCustomEvent(e event_log.Event) {
	o.History = append(o.History, &e)
	GetOrderPersistor().UpsertOrder(o) // Error is ignored because it gets already logged in UpsertOrder()

	jsonBytes, _ := json.MarshalIndent(&e, "", "	")
	event_log.Debug("Saved Order Event! ", string(jsonBytes))
}

// GetCustomer
func (o *Order) GetCustomer(customCustomer interface{}) (c *customer.Customer, err error) {
	c = &customer.Customer{
		Custom: customCustomer,
	}
	// do mongo magic to load customCustomer
	return
}

func (o *Order) Insert() error {
	return GetOrderPersistor().InsertOrder(o)
}

func (o *Order) Upsert() error {
	return GetOrderPersistor().UpsertOrder(o)
}

// Convenience method for the default case of adding a position with following upsert in db
func (order *Order) AddPosition(pos *Position) error {
	return order.AddPositionAndUpsert(pos, true)
}

/* Add Position to Order. Use upsert=false when adding multiple positions. Upsert only once when adding last position for better performacne  */
func (o *Order) AddPositionAndUpsert(pos *Position, upsert bool) error {
	existingPos := o.GetPositionByItemId(pos.ItemID)
	if existingPos != nil {
		err := errors.New("position already exists use SetPositionQuantity or GetPositionById to manipulate it")
		o.SaveOrderEventDetailed(ActionAddPosition, err, pos.ItemID, "")
		return err
	}
	o.Positions = append(o.Positions, pos)

	//comment := ""
	if upsert {
		if err := GetOrderPersistor().UpsertOrder(o); err != nil {
			o.SaveOrderEventDetailed(ActionAddPosition, err, pos.ItemID, "Error while adding position to order. Could not upsert order")
			return err
		}
	} else {
		// TODO log if upsert was skipped
		//comment = "Did not perform upsert"
		//log.Println(comment)
	}
	//o.SaveOrderEventDetailed(ActionAddPosition, nil, pos.ItemID, comment)

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
	if err := GetOrderPersistor().UpsertOrder(o); err != nil {
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

func (o *Order) ReportErrors(printOnConsole bool) string {
	errCount := 0
	if len(o.History) > 0 {
		errCount++
		jsonBytes, err := json.MarshalIndent(o.History, "", "	")
		if err != nil {
			panic(err)
		}
		s := string(jsonBytes)
		if printOnConsole {
			log.Println("Errors logged for order with orderID:")
			log.Println(s)
		}

		return s
	}
	return "No errors logged for order with orderID " + o.OrderID
}
