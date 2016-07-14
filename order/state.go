package order

import "github.com/foomo/shop/state"

const (
	StateType            string = "OrderStatus"
	OrderStatusInvalid   string = "orderStatusInvalid"
	OrderStatusCart      string = "orderStatusCart"
	OrderStatusConfirmed string = "orderStatusConfirmed"
	OrderStatusComplete  string = "orderStatusComplete"
	OrderStatusCanceled  string = "orderStatusCanceled"
)

var transitions = map[string][]string{
	OrderStatusInvalid:   []string{state.WILDCARD},
	OrderStatusCart:      []string{OrderStatusConfirmed, OrderStatusInvalid},
	OrderStatusConfirmed: []string{OrderStatusComplete, OrderStatusInvalid},
	OrderStatusComplete:  []string{},
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
		Description: "Order has been confirmed and is being processed.",
		Initial:     false,
	},
	OrderStatusComplete: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusComplete,
		Description: "Order has been completed.",
		Initial:     false,
	},
}

var DefaultStateMachine = &state.StateMachine{
	InitialState: OrderStatusCart,
	Transitions:  transitions,
	BluePrints:   blueprints,
}
