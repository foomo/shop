package order

import (
	"errors"
	"time"

	"github.com/foomo/shop/customer"
	"github.com/foomo/shop/payment"
	"github.com/foomo/shop/shipping"
	"github.com/foomo/shop/state"
	"github.com/foomo/shop/utils"
	"github.com/foomo/shop/version"
)

//------------------------------------------------------------------
// ~ SIMPLE GETTERS ON ORDER
//------------------------------------------------------------------

func (order *Order) GetID() string {
	return order.Id
}

func (order *Order) GetVersion() *version.Version {
	return order.Version
}

// GetCustomerId
func (order *Order) GetCustomerId() string {
	return order.CustomerId
}

func (order *Order) GetOrderType() OrderType {
	return order.OrderType
}

// GetCustomer returns the latest version of the customer or version specified in CustomerFreeze
func (order *Order) GetCustomer(customProvider customer.CustomerCustomProvider) (c *customer.Customer, err error) {
	if order.IsFrozenCustomer() {
		return customer.GetCustomerByVersion(order.CustomerId, order.CustomerFreeze.Version, customProvider)
	}
	return customer.GetCustomerById(order.CustomerId, customProvider)
}

func (order *Order) GetPositions() []*Position {
	return order.Positions
}

func (order *Order) GetAddressBillingId() string {
	return order.AddressBillingId
}

func (order *Order) GetAddressShippingId() string {
	return order.AddressShippingId
}

func (order *Order) GetPayment() *payment.Payment {
	return order.Payment
}
func (order *Order) GetShipping() *shipping.ShippingProperties {
	return order.Shipping
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
func (order *Order) GetCompletedAt() time.Time {
	return order.CompletedAt
}
func (order *Order) GetCompletedAtFormatted() string {
	return utils.GetFormattedTime(order.CompletedAt)
}

// func (order *Order) GetStatus() OrderStatus {
// 	return order.Status
// }
func (order *Order) GetState() *state.State {
	return order.StateWrapper.State
}

func (order *Order) GetStateMachine() (*state.StateMachine, error) {
	stateMachine, ok := StateMachineMap[order.StateWrapper.StateMachineId]
	if !ok {
		return nil, errors.New("No StateMachine available for id: " + order.StateWrapper.StateMachineId)
	}
	return stateMachine, nil
}

//------------------------------------------------------------------
// ~ SIMPLE SETTERS ON ORDER
//------------------------------------------------------------------

func (order *Order) SetStateMachine(stateMachineId string) error {
	order.StateWrapper.StateMachineId = stateMachineId
	return order.Upsert()
}

func (order *Order) SetState(targetState string) error {
	return order.setState(targetState, false)
}
func (order *Order) ForceState(targetState string) error {
	return order.setState(targetState, true)
}
func (order *Order) setState(targetState string, force bool) error {
	stateMachine, err := order.GetStateMachine()
	if err != nil {
		return err
	}
	var state *state.State
	//var err error

	if force {
		state, err = stateMachine.ForceTransitionToState(order.GetState(), targetState)
	} else {
		state, err = stateMachine.TransitionToState(order.GetState(), targetState)
	}

	if err != nil {
		return err
	}
	order.StateWrapper.State = state
	return order.Upsert()
}
func (order *Order) SetCompleted() error {
	order.CompletedAt = utils.TimeNow()
	return order.Upsert()
}
func (order *Order) SetModified() error {
	order.LastModifiedAt = utils.TimeNow()
	return order.Upsert()
}
func (order *Order) SetShipping(shipping *shipping.ShippingProperties) error {
	order.Shipping = shipping
	return order.Upsert()
}
func (order *Order) SetPayment(payment *payment.Payment) error {
	order.Payment = payment
	return order.Upsert()
}

func (order *Order) SetPositions(positions []*Position) error {
	order.Positions = positions
	return order.Upsert()
}

// TODO this does not check if id exists
func (order *Order) SetAddressShippingId(id string) error {
	if order.IsFrozenCustomer() {
		return errors.New("Error: Shipping Address cannot be changed after customer freeze.")
	}
	order.AddressShippingId = id
	return order.Upsert()
}

// TODO this does not check if id exists
func (order *Order) SetAddressBillingId(id string) error {
	if order.IsFrozenCustomer() {
		return errors.New("Error: Shipping Address cannot be changed after customer freeze.")
	}
	order.AddressBillingId = id
	return order.Upsert()
}

func (order *Order) SetCustomerId(id string) error {
	if order.IsFrozenCustomer() {
		return errors.New("Error: CustomerId cannot be changed after customer freeze.")
	}
	order.CustomerId = id
	return order.Upsert()
}

func (order *Order) SetOrderType(t OrderType) error {
	order.OrderType = t
	return order.Upsert()
}
