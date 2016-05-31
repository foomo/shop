package order

import "github.com/foomo/shop/state"

const (
	StateType            string = "OrderStatus"
	OrderStatusInvalid   string = "orderStatusInvalid"
	OrderStatusCreated   string = "orderStatusCreated"
	OrderStatusConfirmed string = "orderStatusConfirmed"
	OrderStatusComplete  string = "orderStatusComplete"
	OrderStatusCanceled  string = "orderStatusCanceled"
	DefaultStateMachine  string = "defaultStateMachine"
)

var transitions = map[string][]string{
	OrderStatusInvalid:   []string{state.WILDCARD},
	OrderStatusCreated:   []string{OrderStatusConfirmed, OrderStatusInvalid},
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
	OrderStatusCreated: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusCreated,
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

var defaultStateMachine = &state.StateMachine{
	InitialState: OrderStatusCreated,
	Transitions:  transitions,
	BluePrints:   blueprints,
}

var StateMachineMap map[string]*state.StateMachine = map[string]*state.StateMachine{
	DefaultStateMachine: defaultStateMachine,
}
