package event_log

import (
	"encoding/json"
	"log"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Event struct {
	ID             bson.ObjectId `bson:"_id,omitempty"`
	Type           EventType
	Action         string
	OrderID        string
	PositionItemID string
	Comment        string
	Error          string // not type error, because jsonMarshal does not work on error
	Timestamp      time.Time
	Custom         interface{}
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
	ActionInsertingOrder      ActionShop = "actionInsertingOrder"
	ActionUpsertingOrder      ActionShop = "actionUpsertingOrder"
	ActionCreateOrder         ActionShop = "actionCreatingOrder"
	ActionDropEventCollection ActionShop = "actionDropEventCollection"
	ActionDropAllCollections  ActionShop = "actionDropAllCollections"
	ActionValidate            ActionShop = "actionValidate"
	ActionRetrieveOrder       ActionShop = "actionRetrieveOrder"
	ActionStatusUpdate        ActionShop = "actionStatusUpdate"
	ActionApplyATPResponse    ActionShop = "actionApplyATPResponse"
	ActionSendATPRequest      ActionShop = "actionSendATPRequest"
	ActionSendOrder           ActionShop = "actionSendOrder"
)

// EventHistory is a field of Order
type EventHistory []*Event

func (eh EventHistory) Report() string {
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

// Print all shop errors in console
func ReportShopErrors() {
	p := GetEventPersistor()

	event := &Event{}
	q := p.GetCollection().Find(&bson.M{"type": EventTypeError})
	iter := q.Iter()
	var errCount int
	for iter.Next(event) {
		errCount++
		jsonBytes, err := json.MarshalIndent(event, "", "	")
		if err != nil {
			panic(err)
		}
		log.Println(string(jsonBytes))
	}
	log.Println("Errors:", errCount)
}

// TODO does not work, errors do not show up
// Print all Errors on Orders in console
func ReportOrderErrors() {
	p := GetEventPersistor()

	eventHistory := &EventHistory{}
	q := p.GetCollection().Find(&bson.M{}).Select(bson.M{"eventhistory": true})
	iter := q.Iter()
	var errCount int
	for iter.Next(eventHistory) {
		if len(*eventHistory) > 0 {
			errCount++
			jsonBytes, err := json.MarshalIndent(eventHistory, "", "	")
			if err != nil {
				panic(err)
			}
			log.Println(string(jsonBytes))
		}
	}
	log.Println("Errors:", errCount)

}
