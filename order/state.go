package order

import "github.com/foomo/shop/state"

const (
	StateType              string = "OrderStatus"
	OrderStatusInvalid     string = "OrderStatusInvalid"
	OrderStatusCart        string = "OrderStatusCart"
	OrderStatusConfirmed   string = "OrderStatusConfirmed"
	OrderStatusTransmitted string = "OrderStatusTransmitted"
	OrderStatusShipped     string = "OrderStatusShipped"
	OrderStatusComplete    string = "OrderStatusComplete"
	OrderStatusCanceled    string = "OrderStatusCanceled"
)

var transitions = map[string][]string{
	OrderStatusInvalid:     []string{state.WILDCARD},
	OrderStatusCart:        []string{OrderStatusConfirmed, OrderStatusInvalid},
	OrderStatusConfirmed:   []string{OrderStatusTransmitted, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusTransmitted: []string{OrderStatusShipped, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusShipped:     []string{OrderStatusComplete, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusComplete:    []string{},
}

// blueprints for possible states
var blueprints = map[string]state.BluePrint{
	OrderStatusInvalid: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusInvalid,
		Description: "Something went wrong",
		Initial:     false,
	},
	OrderStatusCart: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusCart,
		Description: "Order has been created.",
		Initial:     true,
	},
	OrderStatusConfirmed: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusConfirmed,
		Description: "Order has been confirmed by the Webshop.",
		Initial:     false,
	},
	OrderStatusTransmitted: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusTransmitted,
		Description: "Order has been transmitted to external system.",
		Initial:     false,
	},
	OrderStatusShipped: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusShipped,
		Description: "Order has been shipped.",
		Initial:     false,
	},
	OrderStatusComplete: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusComplete,
		Description: "Order has been completed.",
		Initial:     false,
	},
	OrderStatusCanceled: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusCanceled,
		Description: "Order has been canceled.",
		Initial:     false,
	},
}

var DefaultStateMachine = &state.StateMachine{
	InitialState: OrderStatusCart,
	Transitions:  transitions,
	BluePrints:   blueprints,
}
