package event_log

import (
	"encoding/json"
	"log"
	"time"

	"github.com/foomo/shop/debug"
	"github.com/foomo/shop/trace"
	"github.com/foomo/shop/utils"

	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type Event struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Type      EventType     // Success or Error. This is not set explicitely but derived from err == nil or err != nil
	Action    string
	Info      *Info
	Error     string // not type error, because jsonMarshal does not work on error
	Timestamp time.Time
}

type Info struct {
	Description string
	Caller      trace.Callers `bson:",omitempty"`
	OrderId     string        `bson:",omitempty"`
	CustomerId  string        `bson:",omitempty"`
}

type EventType string

type ActionShop string

//------------------------------------------------------------------
// ~ CONSTANTS & VARS
//------------------------------------------------------------------
const (
	EventTypeSuccess          EventType  = "EventSuccess"
	EventTypeError            EventType  = "EventError"
	ActionTest                ActionShop = "actionTest"
	ActionCreateOrder         ActionShop = "actionCreateOrder"
	ActionRetrieveOrder       ActionShop = "actionRetrieveOrder"
	ActionUpsertingOrder      ActionShop = "actionUpsertOrder"
	ActionDeleteOrder         ActionShop = "actionDeleteOrder"
	ActionCreateCustomer      ActionShop = "actionCreateCustomer"
	ActionRetrieveCustomer    ActionShop = "actionRetrieveCustomer"
	ActionUpsertingCustomer   ActionShop = "actionUpsertCustomer"
	ActionDeleteCustomer      ActionShop = "actionDeleteCustomer"
	ActionDropEventCollection ActionShop = "actionDropEventCollection"
	ActionDropAllCollections  ActionShop = "actionDropAllCollections"
	ActionValidate            ActionShop = "actionValidate"
	ActionStatusUpdate        ActionShop = "actionStatusUpdate"
	ActionApplyATPResponse    ActionShop = "actionApplyATPResponse"
	ActionSendATPRequest      ActionShop = "actionSendATPRequest"
	ActionSendOrder           ActionShop = "actionSendOrder"
)

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

func NewEvent() *Event {
	return &Event{
		Info:      &Info{},
		Timestamp: utils.TimeNow(),
	}
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func SaveShopEvent(action ActionShop, info *Info, err error, description string) {
	debug.Log("Action", string(action))
	event := NewEvent()
	if info != nil {
		event.Info = info
	}

	event.Info.Caller = trace.WhoCalledMe(4)
	if err != nil {
		event.Type = EventTypeError
		event.Error = err.Error()
	} else {
		event.Type = EventTypeSuccess
	}

	event.Action = string(action)
	event.Info.Description = description

	if !saveShopEventDB(event) {
		jsonBytes, err := json.MarshalIndent(event, "", "	")
		if err != nil {
			log.Println("Could not jsonMarshal event")
		}
		log.Println("Saving Shop Event failed! ", string(jsonBytes))
	}
	jsonBytes, _ := json.MarshalIndent(event, "", "	")
	debug.Log("Saved Shop Event! ", string(jsonBytes))
}

//------------------------------------------------------------------
// ~ Report Methods - no returns, only prints
//------------------------------------------------------------------

func Report(query *bson.M) {
	p := GetEventPersistor()

	event := &Event{}
	q := p.GetCollection().Find(query)
	iter := q.Iter()
	var errCount int
	for iter.Next(event) {
		if event.Type == EventTypeError {
			errCount++
		}
		jsonBytes, err := json.MarshalIndent(event, "", "	")
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(string(jsonBytes))
	}
	log.Println("Errors:", errCount)
}

// Print all shop events in console
func ReportShopEvents() {
	Report(&bson.M{})
}

// Print all shop error events in console
func ReportShopErrors() {
	Report(&bson.M{"type": EventTypeError})
}

// Print all shop success events in console
func ReportShopSuccess() {
	Report(&bson.M{"type": EventTypeSuccess})
}

func ResetShopEventLog() bool {
	err := GetEventPersistor().GetCollection().DropCollection()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//------------------------------------------------------------------
// ~ String Methods - get events as formatted string
//------------------------------------------------------------------

func GetAllShopEventsAsString() string {
	return GetShopEventsAsString(&bson.M{})
}

func GetShopErrorsAsString() string {
	return GetShopEventsAsString(&bson.M{"type": EventTypeError})
}

func LogShopEvents() {
	log.Println(GetAllShopEventsAsString())
}

func LogShopErrors() {
	log.Println(GetShopErrorsAsString())
}

func GetShopEventsAsString(query *bson.M) string {
	p := GetEventPersistor()
	var result string

	iter := p.GetCollection().Find(query).Iter()
	event := &Event{}
	for iter.Next(event) {
		jsonBytes, err := json.MarshalIndent(event, "", "	")
		if err != nil {
			continue
		}
		result += string(jsonBytes)
	}
	return result
}
