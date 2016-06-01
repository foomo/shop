// Package order handles order in a shop
// - carts are incomplete orders
//
package order

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/payment"
	"github.com/foomo/shop/shipping"
	"github.com/foomo/shop/state"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

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
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type ActionOrder string
type OrderType string
type OrderStatus string

// Order of item
// create revisions
type Order struct {
	BsonId            bson.ObjectId `bson:"_id,omitempty"`
	Id                string        // automatically generated unique id
	Version           *version.Version
	unlinkDB          bool // if true, changes to Customer are not stored in database
	Flags             *Flags
	StateWrapper      *state.StateWrapper
	CustomerId        string
	CustomerFreeze    *Freeze
	AddressBillingId  string
	AddressShippingId string
	OrderType         OrderType
	CreatedAt         time.Time
	LastModifiedAt    time.Time
	CompletedAt       time.Time
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

type Flags struct {
	forceUpsert bool
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
	ItemID       string
	StateWrapper *state.StateWrapper
	Name         string
	Description  string
	Quantity     float64
	QuantityUnit string
	Price        float64
	IsATPApplied bool
	Refund       bool
	Custom       interface{}
}

type Freeze struct {
	Version int
	Time    time.Time
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewOrder creates a new Order in the database and returns it.
func NewOrder(customProvider OrderCustomProvider) (*Order, error) {
	return NewOrderWithCustomId(customProvider, nil)
}

// NewOrderWithCustomId creates a new Order in the database and returns it.
// With orderIdFunc, an optional method can be specified to generate the orderId. If nil, a default algorithm is used.
func NewOrderWithCustomId(customProvider OrderCustomProvider, orderIdFunc func() (string, error)) (*Order, error) {
	var orderId string
	if orderIdFunc != nil {
		var err error
		orderId, err = orderIdFunc()
		if err != nil {
			return nil, err
		}
	} else {
		orderId = unique.GetNewID()
	}
	order := &Order{
		StateWrapper: &state.StateWrapper{
			StateMachineId: DefaultStateMachine,
		},
		Flags:          &Flags{},
		Id:             orderId,
		Version:        version.NewVersion(),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		CustomerFreeze: &Freeze{},
		OrderType:      OrderTypeOrder,
		Positions:      []*Position{},
		Payment:        &payment.Payment{},
		PriceInfo:      &OrderPriceInfo{},
		Shipping:       &shipping.ShippingProperties{},
	}

	stateMachine, err := order.GetStateMachine()
	if err != nil {
		return nil, err
	}
	order.StateWrapper.State = stateMachine.GetInitialState()

	if customProvider != nil {
		order.Custom = customProvider.NewOrderCustom()
	}

	// Store order in database
	err = order.insert()
	// Retrieve order again from. (Otherwise upserts on order would fail because of missing mongo ObjectID)
	order, err = GetOrderById(order.Id, customProvider)
	return order, err

}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON ORDER
//------------------------------------------------------------------

// Unlinks order from database. No peristent changes are performed until order is linked again.
func (order *Order) UnlinkFromDB() {
	order.unlinkDB = true
}
func (order *Order) LinkDB() {
	order.unlinkDB = false
}

// Returns true, if order is associated to a Customer id.
// Otherwise the order is a cart of on anonymous user
func (order *Order) HasCustomer() bool {
	return order.CustomerId != ""
}

// Insert order into database
func (order *Order) insert() error {
	return insertOrder(order)
}
func (order *Order) Upsert() error {
	return UpsertOrder(order)
}
func (order *Order) UpsertAndGetOrder(customProvider OrderCustomProvider) (*Order, error) {
	return UpsertAndGetOrder(order, customProvider)
}
func (order *Order) Delete() error {
	return DeleteOrder(order)
}

// FreezeCustomer associates the current version of the customer with the order
// Changes on customer after freeze are no longer considered for this order.
func (order *Order) FreezeCustomer() error {
	if order.IsFrozenCustomer() {
		return errors.New("Customer version has already been frozen! Use UnfreezeCustomer() is necessary")
	}
	if !order.HasCustomer() {
		return errors.New("No customer is associated to this order yet!")
	}
	customer, err := order.GetCustomer(nil)
	if err != nil {
		return err
	}
	order.CustomerFreeze = &Freeze{
		Version: customer.GetVersion().Current,
		Time:    utils.TimeNow(),
	}
	return nil
}
func (order *Order) UnFreezeCustomer() {
	order.CustomerFreeze = &Freeze{}
}
func (order *Order) IsFrozenCustomer() bool {
	return !order.CustomerFreeze.Time.IsZero()
}

// Add Position to Order.
func (order *Order) AddPosition(pos *Position) error {
	existingPos := order.GetPositionByItemId(pos.ItemID)
	if existingPos != nil {
		err := errors.New("position already exists use SetPositionQuantity or GetPositionById to manipulate it")
		//order.SaveOrderEvent(ActionAddPosition, err, "Position: "+pos.ItemID)
		return err
	}
	order.Positions = append(order.Positions, pos)

	if err := order.Upsert(); err != nil {
		//	description := "Could not add position " + pos.ItemID + ".  Upsert failed"
		//	order.SaveOrderEvent(ActionAddPosition, err, description)
		return err
	}

	return nil
}

func (order *Order) SetPositionQuantity(itemID string, quantity float64) error {
	pos := order.GetPositionByItemId(itemID)
	if pos == nil {
		err := fmt.Errorf("position with %q not found in order", itemID)
		//order.SaveOrderEvent(ActionChangeQuantityPosition, err, "Could not set quantity of position "+pos.ItemID+" to "+fmt.Sprint(quantity))
		return err
	}
	pos.Quantity = quantity
	//order.SaveOrderEvent(ActionChangeQuantityPosition, nil, "Set quantity of position "+pos.ItemID+" to "+fmt.Sprint(quantity))
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
		//order.SaveOrderEvent(ActionChangeQuantityPosition, err, "Could not update quantity for position "+pos.ItemID+". Upsert failed.")
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

// OverrideID may be used to use a different than the automatially generated if
func (order *Order) OverrideId(id string) error {
	order.Id = id
	return order.Upsert()
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON POSITION
//------------------------------------------------------------------

func (p *Position) IsRefund() bool {
	return p.Refund
}

// GetAmount returns the Price Sum of the position
func (p *Position) GetAmount() float64 {
	return p.Price * p.Quantity
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// DiffTwoLatestOrderVersions compares the two latest Versions of Order found in version.
// If openInBrowser, the result is automatically displayed in the default browser.
func DiffTwoLatestOrderVersions(orderId string, customProvider OrderCustomProvider, openInBrowser bool) (string, error) {
	version, err := GetCurrentVersionOfOrderFromVersionsHistory(orderId)
	if err != nil {
		return "", err
	}

	return DiffOrderVersions(orderId, version.Current-1, version.Current, customProvider, openInBrowser)
}

func DiffOrderVersions(orderId string, versionA int, versionB int, customProvider OrderCustomProvider, openInBrowser bool) (string, error) {
	if versionA <= 0 || versionB <= 0 {
		return "", errors.New("Error: Version must be greater than 0")
	}
	name := "order_v" + strconv.Itoa(versionA) + "_vs_v" + strconv.Itoa(versionB)
	orderVersionA, err := GetOrderByVersion(orderId, versionA, customProvider)
	if err != nil {
		return "", err
	}
	orderVersionB, err := GetOrderByVersion(orderId, versionB, customProvider)
	if err != nil {
		return "", err
	}

	html, err := version.DiffVersions(orderVersionA, orderVersionB)
	if err != nil {
		return "", err
	}
	if openInBrowser {
		err := utils.OpenInBrowser(name, html)
		if err != nil {
			log.Println(err)
		}
	}
	return html, err
}

// add a global State Machine. Don't forget to set the id on the objects that are
// supposed to use this State Machine
func AddStateMachine(stateMachineId string, stateMachine *state.StateMachine) {
	StateMachineMap[stateMachineId] = stateMachine
}
