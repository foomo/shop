package event_log

import (
	"encoding/json"
	"log"
	"time"
)

type Event struct {
	Type      EventType
	Action    string
	OrderID   string
	Comment   string
	Error     string // not type error, because jsonMarshal does not work on error
	Timestamp time.Time
	Custom    interface{}
}

func NewEvent() *Event {
	return &Event{
		Timestamp: time.Now(),
	}
}

type EventType string

const (
	EventTypeSuccess EventType = "EventSuccess"
	EventTypeError   EventType = "EventError"
)

type ActionShop string

const (
	ActionInsertingOrder   ActionShop = "actionInsertingOrder"
	ActionUpsertingOrder   ActionShop = "actionUpsertingOrder"
	ActionCreateOrder      ActionShop = "actionCreatingOrder"
	ActionDropCollection   ActionShop = "actionDropCollection"
	ActionValidate         ActionShop = "actionValidate"
	ActionRetrieveOrder    ActionShop = "actionRetrieveOrder"
	ActionStatusUpdate     ActionShop = "actionStatusUpdate"
	ActionApplyATPResponse ActionShop = "actionApplyATPResponse"
	ActionSendATPRequest   ActionShop = "actionSendATPRequest"
	ActionSendOrder   ActionShop = "actionSendOrder"
)

type EventHistory []*Event

func (eh *EventHistory) Report() string {
	jsonBytes, err := json.MarshalIndent(eh, "", "	")
	if err != nil {
		log.Println("Could not parse json")
		return ""
	}
	return string(jsonBytes)
}

func (eh EventHistory) ReportErrors() string {
	var result string
	for _, e := range eh {
		if e.Type == EventTypeError {
			jsonBytes, err := json.MarshalIndent(e, "", "	")
			if err != nil {
				log.Println("Could not parse json")
				return ""
			}
			result += string(jsonBytes)
		}
	}
	return result
}

func Debug(i ...interface{}) {
	if VERBOSE {
		log.Println("[DEBUG]", i)
	}
}
