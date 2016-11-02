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

	"github.com/foomo/shop/address"
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
	ActionStatusUpdateHead       ActionOrder  = "actionStatusUpdateHead"
	ActionStatusUpdatePosition   ActionOrder  = "actionStatusUpdatePosition"
	ActionNoATPResponseForItemID ActionOrder  = "actionNoATPResponseForItemID"
	ActionValidateStatusHead     ActionOrder  = "actionValidateStatusHead"
	ActionValidateStatusPosition ActionOrder  = "actionValidateStatusPosition"
	ActionAddPosition            ActionOrder  = "actionAddPosition"
	ActionRemovePosition         ActionOrder  = "actionRemovePosition"
	ActionChangeQuantityPosition ActionOrder  = "actionChangeQuantityPosition"
	ActionCreateCustomOrder      ActionOrder  = "actionCreateCustomOrder"
	ActionValidation             ActionOrder  = "actionValidation"
	OrderTypeOrder               OrderType    = "order"
	OrderTypeReturn              OrderType    = "return"
	LanguageCodeGerman           LanguageCode = "DE"
	LanguageCodeFrench           LanguageCode = "FR"
	LanguageCodeItalian          LanguageCode = "IT"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type ActionOrder string
type OrderType string
type OrderStatus string
type LanguageCode string

// Order of item
// create revisions
type Order struct {
	BsonId           bson.ObjectId `bson:"_id,omitempty"`
	CartId           string        // unique cartId. This is the initial id when the cart is created
	Id               string        // unique orderId. This is set when the order is confirmed and sent
	Version          *version.Version
	referenceVersion int  // Version of final order as it was submitted by customer
	unlinkDB         bool // if true, changes to Customer are not stored in database
	Flags            *Flags
	State            *state.State
	CustomerFreeze   *Freeze
	CustomerData     *CustomerData
	OrderType        OrderType
	CreatedAt        time.Time
	ConfirmedAt      time.Time // not yet used
	LastModifiedAt   time.Time
	CompletedAt      time.Time
	Positions        []*Position
	Payment          *payment.Payment
	PriceInfo        *OrderPriceInfo
	Shipping         *shipping.ShippingProperties
	LanguageCode     LanguageCode
	CustomProvider   OrderCustomProvider
	Custom           interface{} `bson:",omitempty"`
}

type CustomerData struct {
	CustomerId      string
	Email           string
	BillingAddress  *address.Address
	ShippingAddress *address.Address
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
	//Fields() *bson.M
}

// Position in an order
type Position struct {
	ItemID       string
	State        *state.State
	Name         string
	Description  string
	Quantity     float64
	QuantityUnit string
	Price        float64
	CrossPrice   float64
	IsATPApplied bool
	IsShipping   bool
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
		State:          DefaultStateMachine.GetInitialState(),
		Flags:          &Flags{},
		CartId:         unique.GetNewID(),
		Id:             orderId,
		Version:        version.NewVersion(),
		CreatedAt:      utils.TimeNow(),
		LastModifiedAt: utils.TimeNow(),
		CustomerFreeze: &Freeze{},
		CustomerData:   &CustomerData{},
		OrderType:      OrderTypeOrder,
		Positions:      []*Position{},
		Payment:        &payment.Payment{},
		PriceInfo:      &OrderPriceInfo{},
		Shipping:       &shipping.ShippingProperties{},
	}

	if customProvider != nil {
		order.Custom = customProvider.NewOrderCustom()
	}

	// Store order in database
	err := order.insert()
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
	return order.CustomerData.CustomerId != ""
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
	if order.CustomerFreeze == nil {
		return false
	}
	return !order.CustomerFreeze.Time.IsZero()
}

// ReplacePosition replaces the itemId of a position, e.g. if article is desired with a different size or color. Quantity is preserved.
func (order *Order) ReplacePosition(itemIdCurrent, itemIdNew string, crossPrice float64, price float64, customProvider OrderCustomProvider) error {
	pos := order.GetPositionByItemId(itemIdCurrent)
	if pos == nil {
		err := fmt.Errorf("position with %q not found in order", itemIdCurrent)
		return err
	}
	currentQty := pos.Quantity
	// If the item already exists, teh quantity is accumulated.
	// (Example: The same shirt in two different sizes is in the order. The size of one shirt is changed to the size of the other shirt. Now there would be two positions for the same shirt)
	posMatchNew := order.GetPositionByItemId(itemIdNew)
	if posMatchNew != nil {
		// Remove position for itemIdCurrent
		err := order.SetPositionQuantity(itemIdCurrent, 0, -1, .1, customProvider)
		if err != nil {
			return err
		}
		// And adjust quantity for already existing position for itemIdNew
		return order.SetPositionQuantity(itemIdNew, currentQty+posMatchNew.Quantity, -1, -1, customProvider)
	}

	// Otherwise replace current position
	pos.ItemID = itemIdNew
	pos.CrossPrice = crossPrice
	pos.Price = price

	return order.Upsert()
}

// Increase Quantity by one. Price is required, if item is not already part of order
func (order *Order) IncPositionQuantity(itemID string, crossPrice float64, price float64, customProvider OrderCustomProvider) error {
	pos := order.GetPositionByItemId(itemID)
	quantity := 1.0
	if pos != nil {
		quantity = pos.Quantity + 1
	}
	return order.SetPositionQuantity(itemID, quantity, crossPrice, price, customProvider)
}

func (order *Order) AddToPositionQuantity(itemID string, addQty float64, crossPrice float64, price float64, customProvider OrderCustomProvider) error {
	pos := order.GetPositionByItemId(itemID)
	quantity := 1.0
	if pos != nil {
		quantity = pos.Quantity + addQty
	}
	return order.SetPositionQuantity(itemID, quantity, crossPrice, price, customProvider)
}
func (order *Order) DecPositionQuantity(itemID string, crossPrice float64, price float64, customProvider OrderCustomProvider) error {
	pos := order.GetPositionByItemId(itemID)
	if pos == nil {
		err := fmt.Errorf("position with %q not found in order", itemID)
		return err
	}
	return order.SetPositionQuantity(itemID, pos.Quantity-1, crossPrice, price, customProvider)
}

// Add Position to Order.
func (order *Order) AddPosition(pos *Position) error {
	existingPos := order.GetPositionByItemId(pos.ItemID)
	if existingPos != nil {
		return nil
	}
	order.Positions = append(order.Positions, pos)

	return order.Upsert()
}

// TODO maybe this is probably the wrong place to set the price
func (order *Order) SetPositionQuantity(itemID string, quantity float64, crossPrice float64, price float64, customProvider OrderCustomProvider) error {
	log.Println("SetPositionQuantity(", itemID, quantity, price, ")")
	pos := order.GetPositionByItemId(itemID)
	// If position for this itemID does not yet exist, create it.
	if pos == nil {
		if quantity > 0 {

			newPos := &Position{
				// TODO initial state is not yet set
				ItemID:   itemID,
				Quantity: quantity,
				Custom:   customProvider.NewPositionCustom(),
			}

			// -1 is used by methods which only change the quantity and not the price
			if crossPrice != -1 {
				newPos.CrossPrice = crossPrice
			}
			if price != -1 {
				newPos.Price = price
			}
			order.Positions = append(order.Positions, newPos)
			return order.Upsert()
		}
		return nil
	}

	// Remove position if quantity is less or equal than zero
	if quantity <= 0.0 {
		positions := []*Position{}
		for _, position := range order.Positions {
			if position.ItemID == itemID {
				continue
			}
			positions = append(positions, position)
		}
		order.Positions = positions
		return order.Upsert()

	} else {
		pos.Quantity = quantity
	}
	// use project-globus-services-1
	// db.orders.find({}, {positions:1}).pretty()
	return order.Upsert()
}
func (order *Order) GetPositionByItemId(itemID string) *Position {
	for _, pos := range order.Positions {
		if pos.ItemID == itemID {
			return pos
		}
	}
	return nil
}

// OverrideID may be used to use a different than the automatially generated id (Unit tests)
func OverrideId(oldID, newID string) error {
	log.Println("+++ INFO: Overriding orderID", oldID, "with id", newID, "+++")
	o := &Order{}
	p := GetOrderPersistor()
	err := p.GetCollection().Find(&bson.M{"id": oldID}).One(o)
	if err != nil {
		log.Println("Upsert failed: Could not find order with id", oldID, "Error:", err)
		return err
	}
	o.Id = newID
	o.State.SetModified()
	_, err = p.GetCollection().UpsertId(o.BsonId, o)
	return err
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS ON POSITION
//------------------------------------------------------------------

func (p *Position) IsRefund() bool {
	return p.Refund
}

// GetAmount returns the Price Sum of the position
func (p *Position) GetPriceTotal() float64 {
	return p.Price * p.Quantity
}
func (p *Position) GetCrossPriceTotal() float64 {
	return p.CrossPrice * p.Quantity
}

func (position *Position) GetState() *state.State {
	return position.State
}

func (position *Position) SetInitialState(stateMachine *state.StateMachine) {
	position.State = stateMachine.GetInitialState()
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
