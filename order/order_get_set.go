package order

import (
	"time"

	"github.com/foomo/shop/customer"
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/payment"
	"github.com/foomo/shop/shipping"
	"github.com/foomo/shop/utils"
)

//------------------------------------------------------------------
// ~ SIMPLE GETTERS ON ORDER
//------------------------------------------------------------------

func (order *Order) GetID() string {
	return order.Id
}

// GetCustomerId
func (order *Order) GetCustomerId() string {
	return order.CustomerId
}

func (order *Order) GetOrderType() OrderType {
	return order.OrderType
}

// GetCustomer
func (order *Order) GetCustomer(customCustomerProvider customer.CustomerCustomProvider) (c *customer.Customer, err error) {
	return customer.GetCustomer(order.CustomerId, customCustomerProvider)
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

func (order *Order) GetHistory() event_log.EventHistory {
	return order.History
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

func (order *Order) GetStatus() OrderStatus {
	return order.Status
}

//------------------------------------------------------------------
// ~ SIMPLE SETTERS ON ORDER
//------------------------------------------------------------------

func (order *Order) SetStatus(status OrderStatus) error {
	order.Status = status
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

// TODO this does not check if id exists
func (order *Order) SetAddressShippingId(id string) error {
	order.AddressShippingId = id
	return order.Upsert()
}
func (order *Order) SetPositions(positions []*Position) error {
	order.Positions = positions
	return order.Upsert()
}

// TODO this does not check if id exists
func (order *Order) SetAddressBillingId(id string) error {
	order.AddressBillingId = id
	return order.Upsert()
}

func (order *Order) SetCustomerId(id string) error {
	order.CustomerId = id
	return order.Upsert()
}

func (order *Order) SetOrderType(t OrderType) error {
	order.OrderType = t
	return order.Upsert()
}
