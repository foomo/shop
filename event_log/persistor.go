package event_log

import (
	"errors"
	"log"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/persistence"
)

//------------------------------------------------------------------
// ~ CONSTANTS & VARS
//------------------------------------------------------------------

var globalEventPersistor *persistence.Persistor

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *persistence.Persistor, err error) {
	return persistence.NewPersistor(mongoURL, collectionName)
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

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

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func saveShopEventDB(e *Event) bool {
	err := GetEventPersistor().GetCollection().Insert(e)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
