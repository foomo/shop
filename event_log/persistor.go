package event_log

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/foomo/shop/configuration"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Persistor struct {
	session        *mgo.Session
	CollectionName string
	db             string
}

const (
	VERBOSE = false
)

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *Persistor, err error) {
	parsedURL, err := url.Parse(mongoURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "mongodb" {
		return nil, fmt.Errorf("missing scheme mongo:// in %q", mongoURL)
	}
	if len(parsedURL.Path) < 2 {
		return nil, errors.New("invalid mongoURL missing db should be mongodb://server:port/db")
	}
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		return nil, err
	}
	p = &Persistor{
		session:        session,
		db:             parsedURL.Path[1:],
		CollectionName: collectionName,
	}
	return
}

func (p *Persistor) GetCollection() *mgo.Collection {
	return p.session.DB(p.db).C(p.CollectionName)
}

func GetPersistor(db string, collection string) *Persistor {
	p, err := NewPersistor(db, collection)
	if err != nil {
		panic(err)
	}
	return p
}

func ResetShopEventLog() bool {
	err := GetPersistor(configuration.MONGO_URL, configuration.MONGO_COLLECTION_SHOP_EVENT_LOG).GetCollection().DropCollection()
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
	p := GetPersistor(configuration.MONGO_URL, configuration.MONGO_COLLECTION_SHOP_EVENT_LOG)
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

func SaveShopEvent(action ActionShop, orderID string, err error) {
	SaveShopEventWithComment(action, orderID, err, "")
}

func SaveShopEventWithComment(action ActionShop, orderID string, err error, comment string) {
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
	event.Comment = comment

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
	err := GetPersistor(configuration.MONGO_URL, configuration.MONGO_COLLECTION_SHOP_EVENT_LOG).InsertEvent(e)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (p *Persistor) InsertEvent(e *Event) error {
	err := p.GetCollection().Insert(e)
	return err
}
