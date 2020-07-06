package order

import (
	"errors"
	"time"

	"github.com/foomo/shop/address"
	"github.com/foomo/shop/customer"
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
func (order *Order) GetReferenceVersion() int {
	return order.referenceVersion
}
func (order *Order) SetReferenceVersion() error {
	if order.referenceVersion != 0 {
		return errors.New("Reference version has already been set and cannot be overridden!")
	}
	order.referenceVersion = order.Version.Current
	return nil
}

// GetCustomerId
func (order *Order) GetCustomerId() string {
	return order.CustomerData.CustomerId
}

// GetCustomer returns the latest version of the customer or version specified in CustomerFreeze
func (order *Order) GetCustomer(customProvider customer.CustomerCustomProvider) (c *customer.Customer, err error) {

	return customer.GetCustomerById(order.CustomerData.CustomerId, customProvider)
}

func (order *Order) GetPositions() []*Position {
	return order.Positions
}

func (order *Order) GetBillingAddress() (*address.Address, error) {
	if order.CustomerData == nil || order.CustomerData.BillingAddress == nil {
		return nil, errors.New("Error: No BillingAddress specified")
	}
	return order.CustomerData.BillingAddress, nil
}
func (order *Order) GetShippingAddress() (*address.Address, error) {
	if order.CustomerData == nil || order.CustomerData.ShippingAddress == nil {
		return nil, errors.New("Error: No BillingAddress specified")
	}
	return order.CustomerData.ShippingAddress, nil
}

// func (order *Order) GetPayment() *payment.Payment {
// 	return order.Payment
// }
// func (order *Order) GetShipping() *shipping.ShippingProperties {
// 	return order.Shipping
// }

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
	return order.State
}

//------------------------------------------------------------------
// ~ SIMPLE SETTERS ON ORDER
//------------------------------------------------------------------

func (order *Order) SetInitialState(stateMachine *state.StateMachine) {
	order.State = stateMachine.GetInitialState()
}

// SetState performs the transition to target state
// If stateMachine is nil, the default state machine is used
func (order *Order) SetState(stateMachine *state.StateMachine, targetState string) error {
	if stateMachine == nil {
		stateMachine = DefaultStateMachine
	}
	err := stateMachine.TransitionToState(order.GetState(), targetState)
	if err != nil {
		return err
	}
	return order.Upsert()
}
func (order *Order) ForceState(stateMachine *state.StateMachine, targetState string) error {
	if stateMachine == nil {
		stateMachine = DefaultStateMachine
	}
	err := stateMachine.ForceTransitionToState(order.GetState(), targetState)
	if err != nil {
		return err
	}
	return order.Upsert()
}
func (order *Order) SetStatePosition(stateMachine *state.StateMachine, targetState string, position *Position) error {
	if stateMachine == nil {
		stateMachine = DefaultStateMachine
	}
	err := stateMachine.TransitionToState(position.GetState(), targetState)
	if err != nil {
		return err
	}
	return order.Upsert()
}
func (order *Order) ForceStatePosition(stateMachine *state.StateMachine, targetState string, position *Position) error {
	if stateMachine == nil {
		stateMachine = DefaultStateMachine
	}
	err := stateMachine.ForceTransitionToState(position.GetState(), targetState)
	if err != nil {
		return err
	}
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

// func (order *Order) SetShipping(shipping *shipping.ShippingProperties) error {
// 	order.Shipping = shipping
// 	return order.Upsert()
// }
// func (order *Order) SetPayment(payment *payment.Payment) error {
// 	order.Payment = payment
// 	return order.Upsert()
// }

func (order *Order) SetPositions(positions []*Position) error {
	order.Positions = positions
	return order.Upsert()
}
func (order *Order) SetPriceOverrideGroup(groupID string) error {
	if order.CustomerData == nil {
		order.CustomerData = &CustomerData{}
	}
	order.CustomerData.PriceOverrideGroup = groupID
	return order.Upsert()
}

func (order *Order) SetBasicCustomerData(id string, addrKey string, customerType string, priceOverrideGroup string) error {

	if order.CustomerData == nil {
		order.CustomerData = &CustomerData{}
	}

	if order.CustomerData.CustomerId == id &&
		order.CustomerData.AddrKey == addrKey &&
		order.CustomerData.CustomerType == customerType &&
		order.CustomerData.PriceOverrideGroup == priceOverrideGroup {
		return nil
	}

	order.CustomerData.CustomerId = id
	order.CustomerData.AddrKey = addrKey
	order.CustomerData.CustomerType = customerType
	order.CustomerData.PriceOverrideGroup = priceOverrideGroup
	return order.Upsert()
}

func (order *Order) SetForceUpsert(force bool) {
	if order.Flags == nil {
		order.Flags = &Flags{}
	}
	order.Flags.forceUpsert = force
	return
}
