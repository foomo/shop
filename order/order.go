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

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/customer"
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/payment"
	"github.com/foomo/shop/shipping"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			CONSTANTS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

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
	OrderTypeOrder               OrderType   = "order"
	OrderTypeReturn              OrderType   = "return"
	OrderStatusInvalid           OrderStatus = "orderStatusInvalid"
	OrderStatusCreated           OrderStatus = "orderStatusCreated"
	OrderStatusPocessed          OrderStatus = "orderStatusProcessed"
	OrderStatusShipped           OrderStatus = "orderStatusShipped"
	OrderStatusReadyForATP       OrderStatus = "orderStatusReadyForATP"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC TYPES
+++++++++++++++++++++++++++++++++++++++++++++++++ */

type ActionOrder string
type OrderType string
type OrderStatus string

// Order of item
// create revisions
type Order struct {
	BsonID            bson.ObjectId `bson:"_id,omitempty"`
	Id                string        // automatically generated unique id
	CustomerId        string
	AddressBillingId  string
	AddressShippingId string
	OrderType         OrderType
	CreatedAt         time.Time
	LastModifiedAt    time.Time
	CompletedAt       time.Time
	Status            OrderStatus
	History           event_log.EventHistory
	Positions         []*Position
	Payment           *payment.Payment
	PriceInfo         *OrderPriceInfo
	Shipping          *shipping.ShippingProperties
	queue             *struct {
		Name           string
		RetryAfter     time.Duration
		LastProcessing time.Time
	}
	Custom interface{} `bson:",omitempty"`
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

type OrderCustomProvider interface {
	NewOrderCustom() interface{}
	NewPositionCustom() interface{}
	Fields() *bson.M
}

// Position in an order
type Position struct {
	//ID           string
	ItemID       string
	Name         string
	Description  string
	Quantity     float64
	QuantityUnit string
	Price        float64
	IsATPApplied bool
	Refund       bool
	Custom       interface{}
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS ON ORDER
+++++++++++++++++++++++++++++++++++++++++++++++++ */

func (order *Order) GetID() string {
	return order.Id
}
func (order *Order) SetStatus(status OrderStatus) {
	order.Status = status
}
func (order *Order) GetStatus() OrderStatus {
	return order.Status
}

func (order *Order) SaveOrderEvent(action ActionOrder, err error, description string) {
	event_log.Debug("Action", string(action), "OrderID", order.Id)
	event := event_log.NewEvent()
	if err != nil {
		event.Type = event_log.EventTypeError
	} else {
		event.Type = event_log.EventTypeSuccess
	}
	event.Action = string(action)
	event.OrderID = order.Id
	event.Description = description
	if err != nil {
		event.Error = err.Error()
	}
	order.History = append(order.History, event)
	order.Upsert() // Error is ignored because it gets already logged in UpsertOrder()

	jsonBytes, _ := json.MarshalIndent(event, "", "	")
	event_log.Debug("Saved Order Event! ", string(jsonBytes))
}

// Event will only be saved if is an error
func (order *Order) SaveOrderEventOnError(action ActionOrder, err error, description string) {
	if err == nil {
		return
	}
	order.SaveOrderEvent(action, err, description)
}

func (order *Order) SaveOrderEventCustomEvent(e event_log.Event) {
	order.History = append(order.History, &e)
	order.Upsert() // Error is ignored because it gets already logged in UpsertOrder()
	jsonBytes, _ := json.MarshalIndent(&e, "", "	")
	event_log.Debug("Saved Order Event! ", string(jsonBytes))
}

// GetCustomerId
func (order *Order) GetCustomerId() string {
	return order.CustomerId
}
func (order *Order) SetCustomerId(id string) {
	order.CustomerId = id
}
func (order *Order) GetOrderType() OrderType {
	return order.OrderType
}
func (order *Order) SetOrderType(t OrderType) {
	order.OrderType = t
}

// GetCustomer
func (order *Order) GetCustomer(customCustomerProvider customer.CustomerCustomProvider) (c *customer.Customer, err error) {
	return customer.GetCustomer(order.CustomerId, customCustomerProvider)
}

func (order *Order) Insert() error {
	return InsertOrder(order) // calls the method defined in persistor.go
}

func (order *Order) Upsert() error {
	return UpsertOrder(order) // calls the method defined in persistor.go
}
func (order *Order) Delete() error {
	return nil // TODO delete order in db
}

// Convenience method for the default case of adding a position with following upsert in db
func (order *Order) AddPosition(pos *Position) error {
	return order.AddPositionAndUpsert(pos, true)
}

/* Add Position to Order. Use upsert=false when adding multiple positions. Upsert only once when adding last position for better performacne  */
func (order *Order) AddPositionAndUpsert(pos *Position, upsert bool) error {
	existingPos := order.GetPositionByItemId(pos.ItemID)
	if existingPos != nil {
		err := errors.New("position already exists use SetPositionQuantity or GetPositionById to manipulate it")
		order.SaveOrderEvent(ActionAddPosition, err, "Position: "+pos.ItemID)
		return err
	}
	order.Positions = append(order.Positions, pos)

	//comment := ""
	if upsert {
		if err := order.Upsert(); err != nil {
			description := "Could not add position " + pos.ItemID + ".  Upsert failed"
			order.SaveOrderEvent(ActionAddPosition, err, description)
			return err
		}
	} else {
		// TODO log if upsert was skipped
		//comment = "Did not perform upsert"
		//log.Println(comment)
	}
	//order.SaveOrderEventDetailed(ActionAddPosition, nil, pos.ItemID, comment)

	return nil
}

func (order *Order) SetPositionQuantity(itemID string, quantity float64) error {
	pos := order.GetPositionByItemId(itemID)
	if pos == nil {
		err := fmt.Errorf("position with %q not found in order", itemID)
		order.SaveOrderEvent(ActionChangeQuantityPosition, err, "Could not set quantity of position "+pos.ItemID+" to "+fmt.Sprint(quantity))
		return err
	}
	pos.Quantity = quantity
	order.SaveOrderEvent(ActionChangeQuantityPosition, nil, "Set quantity of position "+pos.ItemID+" to "+fmt.Sprint(quantity))
	// remove position if quantity is zero
	if pos.Quantity == 0.0 {
		for index := range order.Positions {
			if pos.ItemID == itemID {
				order.Positions = append(order.Positions[:index], order.Positions[index+1:]...)
				return nil
			}
		}
	}
	if err := order.Upsert(); err != nil {
		order.SaveOrderEvent(ActionChangeQuantityPosition, err, "Could not update quantity for position "+pos.ItemID+". Upsert failed.")
		return err
	}
	return nil
}
func (order *Order) GetPositionByItemId(itemID string) *Position {
	for _, pos := range order.Positions {
		if pos.ItemID == itemID {
			return pos
		}
	}
	return nil
}
func (order *Order) GetPositions() []*Position {
	return order.Positions
}
func (order *Order) SetPositions(positions []*Position) {
	order.Positions = positions
}
func (order *Order) ReportErrors(printOnConsole bool) string {
	errCount := 0
	if len(order.History) > 0 {
		errCount++
		jsonBytes, err := json.MarshalIndent(order.History, "", "	")
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
	return "No errors logged for order with orderID " + order.Id
}

// TODO this does not check if id exists
func (order *Order) SetAddressBillingId(id string) error {
	order.AddressBillingId = id
	return nil
}

func (order *Order) GetAddressBillingId() string {
	return order.AddressBillingId
}

// TODO this does not check if id exists
func (order *Order) SetAddressShippingId(id string) error {
	order.AddressShippingId = id
	return nil
}

func (order *Order) GetAddressShippingId() string {
	return order.AddressShippingId
}

func (order *Order) GetHistory() event_log.EventHistory {
	return order.History
}
func (order *Order) GetPayment() *payment.Payment {
	return order.Payment
}
func (order *Order) GetShipping() *shipping.ShippingProperties {
	return order.Shipping
}
func (order *Order) GetQueue() *shipping.ShippingProperties {
	return order.Shipping
}

// OverrideID may be used to use a different than the automatially genrated if
func (order *Order) OverrideId(id string) {
	order.Id = id
}

func (order *Order) GetCreatedAt() time.Time {
	return order.CreatedAt
}
func (order *Order) GetLastModifiedAt() time.Time {
	return order.LastModifiedAt
}
func (order *Order) GetCreatedAtFormatted() string {
	return utils.GetFormattedTime(order.CreatedAt)
}
func (order *Order) GetLastModifiedAtFormatted() string {
	return utils.GetFormattedTime(order.LastModifiedAt)
}

func (order *Order) SetModified() {
	order.LastModifiedAt = utils.TimeNow()
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS ON POSITION
+++++++++++++++++++++++++++++++++++++++++++++++++ */

func (p *Position) IsRefund() bool {
	return p.Refund
}

// GetAmount returns the Price Sum of the position
func (p *Position) GetAmount() float64 {
	return p.Price * p.Quantity
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

// NewOrder
func NewOrder(custom interface{}) *Order {
	order := &Order{
		Id:             unique.GetNewID(),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		Status:         OrderStatusCreated,
		OrderType:      OrderTypeOrder,
		History:        event_log.EventHistory{},
		Positions:      []*Position{},
		Payment:        &payment.Payment{},
		PriceInfo:      &OrderPriceInfo{},
		Shipping:       &shipping.ShippingProperties{},
		Custom:         custom,
	}
	return order
}
