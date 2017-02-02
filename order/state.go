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
	OrderStatusReturn                string = "OrderStatusReturn"
	OrderStatusCanceled              string = "OrderStatusCanceled"
)

var transitions = map[string][]string{
	OrderStatusInvalid:               []string{state.WILDCARD},
	OrderStatusCart:                  []string{OrderStatusConfirmed, OrderStatusInvalid},
	OrderStatusConfirmed:             []string{OrderStatusTransmitted, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusTransmitted:           []string{OrderStatusInProgress, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusInProgress:            []string{OrderStatusPartiallyShipped, OrderStatusShipped, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusShipped:               []string{OrderStatusWaitingForStorePickUp, OrderStatusComplete, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusPartiallyShipped:      []string{OrderStatusWaitingForStorePickUp, OrderStatusShipped, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusWaitingForStorePickUp: []string{OrderStatusComplete, OrderStatusInvalid, OrderStatusCanceled},
	OrderStatusComplete:              []string{OrderStatusReturn},
	OrderStatusReturn:                []string{OrderStatusInvalid},
	OrderStatusCanceled:              []string{OrderStatusInvalid},
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
	OrderStatusInProgress: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusInProgress,
		Description: "Order is being processed.",
		Initial:     false,
	},
	OrderStatusShipped: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusShipped,
		Description: "Order has been shipped.",
		Initial:     false,
	},
	OrderStatusPartiallyShipped: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusPartiallyShipped,
		Description: "Order has been partially shipped.",
		Initial:     false,
	},
	OrderStatusWaitingForStorePickUp: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusWaitingForStorePickUp,
		Description: "Order is ready to be picked up in store.",
		Initial:     false,
	},
	OrderStatusComplete: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusComplete,
		Description: "Order has been completed.",
		Initial:     false,
	},
	OrderStatusReturn: state.BluePrint{
		Type:        StateType,
		Key:         OrderStatusReturn,
		Description: "At least one item of the order has been returned.",
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
