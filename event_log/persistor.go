package event_log

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/persistence"

	"gopkg.in/mgo.v2/bson"
)

const (
	VERBOSE = false
)

var globalEventPersistor *persistence.Persistor

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *persistence.Persistor, err error) {
	return persistence.NewPersistor(mongoURL, collectionName)
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetEventPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_SHOP_EVENT_LOG
	if globalEventPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB global persistor: " + err.Error()))
		}
		globalEventPersistor = p
		return globalEventPersistor
	}

	if url == globalEventPersistor.GetURL() && collection == globalEventPersistor.GetCollectionName() {
		return globalEventPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalEventPersistor = p
	return globalEventPersistor
}

func ResetShopEventLog() bool {
	err := GetEventPersistor().GetCollection().DropCollection()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func GetShopEvents() string {
	return GetShopEventsFromDB(&bson.M{})
}

func GetShopErrors() string {
	return GetShopEventsFromDB(&bson.M{"type": EventTypeError})
}

func LogShopEvents() {
	log.Println(GetShopEvents())
}

func LogShopErrors() {
	log.Println(GetShopErrors())
}

func GetShopEventsFromDB(query *bson.M) string {
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

// func SaveShopEvent(action ActionShop, orderID string, err error) {
// 	SaveShopEventWithComment(action, orderID, err, "")
// }

func SaveShopEvent(action ActionShop, orderID string, err error, description string) {
	Debug("Action", string(action), "OrderID", orderID)
	event := NewEvent()
	if err != nil {
		event.Type = EventTypeError
	} else {
		event.Type = EventTypeSuccess
	}
	event.Action = string(action)
	event.OrderID = orderID
	if err != nil {
		event.Error = err.Error()
	}
	event.Description = description

	if !saveShopEventDB(event) {
		jsonBytes, err := json.MarshalIndent(event, "", "	")
		if err != nil {
			log.Println("Could not jsonMarshal event")
		}
		log.Println("Saving Shop Event failed! ", string(jsonBytes))
	}
	jsonBytes, _ := json.MarshalIndent(event, "", "	")
	Debug("Saved Shop Event! ", string(jsonBytes))
}

func saveShopEventDB(e *Event) bool {
	err := InsertEvent(e)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func InsertEvent(e *Event) error {
	err := GetEventPersistor().GetCollection().Insert(e)
	return err
}
