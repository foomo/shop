package watchlist

import (
	"errors"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/persistence"
)

//------------------------------------------------------------------
// ~ CONSTANTS & VARS
//------------------------------------------------------------------

var globalWatchListPersistor *persistence.Persistor

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
func GetWatchListPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_WATCHLISTS
	if globalWatchListPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB global persistor: " + err.Error()))
		}
		globalWatchListPersistor = p
		return globalWatchListPersistor
	}

	if url == globalWatchListPersistor.GetURL() && collection == globalWatchListPersistor.GetCollectionName() {
		return globalWatchListPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalWatchListPersistor = p
	return globalWatchListPersistor
}

func NewCustomerWatchListsFromCustomerID(customerID string) error {
	return insertCustomerWatchLists(&CustomerWatchLists{
		CustomerID: customerID,
		Lists:      []*WatchList{},
	})
}
func NewCustomerWatchListsFromSessionID(sessionID string) error {
	return insertCustomerWatchLists(&CustomerWatchLists{
		SessionID: sessionID,
		Lists:     []*WatchList{},
	})
}

func DeleteCustomerWatchLists(cw *CustomerWatchLists) error {
	return GetWatchListPersistor().GetCollection().Remove(bson.M{"_id": cw.BsonId})
}
func DeleteCustomerWatchListsByCustomerId(id string) error {
	return GetWatchListPersistor().GetCollection().Remove(bson.M{"customerID": id})
}
func DeleteCustomerWatchListsBySessionId(id string) error {
	return GetWatchListPersistor().GetCollection().Remove(bson.M{"sessionID": id})
}
func DeleteCustomerWatchListsByEmail(email string) error {
	return GetWatchListPersistor().GetCollection().Remove(bson.M{"email": email})
}

func GetCustomerWatchListsByCustomerID(customerID string) (*CustomerWatchLists, error) {
	return findOne(customerID, "", "")
}
func GetCustomerWatchListsBySessionID(sessionID string) (*CustomerWatchLists, error) {
	return findOne("", sessionID, "")
}
func GetCustomerWatchListsByEmail(email string) (*CustomerWatchLists, error) {
	return findOne("", "", email)
}

func (cw *CustomerWatchLists) Upsert() error {
	p := GetWatchListPersistor()
	_, err := p.GetCollection().UpsertId(cw.BsonId, cw)
	return err
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func insertCustomerWatchLists(cw *CustomerWatchLists) error {
	return GetWatchListPersistor().GetCollection().Insert(cw)
}

func findOne(customerID, sessionID, email string) (*CustomerWatchLists, error) {
	if customerID == "" && sessionID == "" && email == "" {
		return nil, errors.New("Error: Either customerID, sessionID or email of Customer must be provided.")
	}
	p := GetWatchListPersistor()
	find := &bson.M{}
	if customerID != "" {
		find = &bson.M{"customerID": customerID}
	}
	if sessionID != "" {
		find = &bson.M{"sessionID": sessionID}
	}
	if customerID != "" {
		find = &bson.M{"email": email}
	}

	CustomerWatchLists := &CustomerWatchLists{}
	err := p.GetCollection().Find(find).One(CustomerWatchLists)
	if err != nil {
		return nil, err
	}

	return CustomerWatchLists, nil
}
