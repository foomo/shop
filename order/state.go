package order

import "github.com/foomo/shop/state"

const (
	StateType                        string = "OrderStatus"
	OrderStatusInvalid               string = "OrderStatusInvalid"
	OrderStatusCart                  string = "OrderStatusCart"
	OrderStatusConfirmed             string = "OrderStatusConfirmed"
	OrderStatusTransmitted           string = "OrderStatusTransmitted"
	OrderStatusInProgress            string = "OrderStatusInProgress"
	OrderStatusPartiallyShipped      string = "OrderStatusPartiallyShipped"
	OrderStatusShipped               string = "OrderStatusShipped"
	OrderStatusWaitingForStorePickUp string = "OrderStatusWaitingForStorePickUp"
	OrderStatusComplete              string = "OrderStatusComplete"
	OrderStatusFullReturn            string = "OrderStatusFullReturn"
	OrderStatusPartialReturn         string = "OrderStatusPartialReturn"
	OrderStatusCanceled              string = "OrderStatusCanceled"
)

var transitions = map[string][]string{
	OrderStatusInvalid:               []string{state.WILDCARD},
	OrderStatusCart:                  []string{OrderStatusConfirmed, OrderStatusInvalid},
	OrderStatusConfirmed:             []string{OrderStatusTransmitted, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusTransmitted:           []string{OrderStatusInProgress, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusInProgress:            []string{OrderStatusPartiallyShipped, OrderStatusShipped, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusShipped:               []string{OrderStatusWaitingForStorePickUp, OrderStatusComplete, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusWaitingForStorePickUp: []string{OrderStatusComplete, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusPartiallyShipped:      []string{OrderStatusWaitingForStorePickUp, OrderStatusShipped, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusComplete:              []string{OrderStatusFullReturn, OrderStatusPartialReturn},
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

func GetStates() map[string]state.BluePrint {
	return blueprints
}

var DefaultStateMachine = &state.StateMachine{
	InitialState: OrderStatusCart,
	Transitions:  transitions,
	BluePrints:   blueprints,
}
