package watchlist

import (
	"errors"
	"log"

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
	url := configuration.GetMongoURL()
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

func NewCustomerWatchListsFromCustomerID(customerID string) (*CustomerWatchLists, error) {
	exists, err := CustomerWatchListsExists(customerID, "", "")
	if err != nil {
		return nil, err
	}
	if exists {
		log.Println("Did not insert CustomerWatchLists for customer", customerID, "- vo already exists")
		return nil, errors.New("CustomerWatchLists for customerID " + customerID + " already exists.")
	}
	err = insertCustomerWatchLists(&CustomerWatchLists{
		CustomerID: customerID,
		Lists:      []*WatchList{},
	})
	if err != nil {
		return nil, err
	}
	return GetCustomerWatchListsByCustomerID(customerID)
}
func NewCustomerWatchListsFromSessionID(sessionID string) (*CustomerWatchLists, error) {
	exists, err := CustomerWatchListsExists("", sessionID, "")
	if err != nil {
		return nil, err
	}
	if exists {
		log.Println("Did not insert CustomerWatchLists for session", sessionID, "- vo already exists")
		return nil, errors.New("CustomerWatchLists for sessionID " + sessionID + " already exists.")
	}
	err = insertCustomerWatchLists(&CustomerWatchLists{
		SessionID: sessionID,
		Lists:     []*WatchList{},
	})
	if err != nil {
		return nil, err
	}
	return GetCustomerWatchListsBySessionID(sessionID)
}

func CustomerWatchListsExists(customerId, sessionId, email string) (bool, error) {


	key := ""
	value := ""
	if customerId != "" {
		key = "customerID"
		value = customerId
	} else if sessionId != "" {
		key = "sessionID"
		value = sessionId
	} else if email != "" {
		key = "email"
		value = email
	}
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	q := collection.Find(&bson.M{key: value})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteCustomerWatchLists(cw *CustomerWatchLists) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	return collection.Remove(bson.M{"_id": cw.BsonId})
}
func DeleteCustomerWatchListsByCustomerId(id string) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	return collection.Remove(bson.M{"customerID": id})
}
func DeleteCustomerWatchListsBySessionId(id string) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	return collection.Remove(bson.M{"sessionID": id})
}
func DeleteCustomerWatchListsByEmail(email string) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	return collection.Remove(bson.M{"email": email})
}

// GetCustomerWatchListsByURIHash returns the CustomerWatchLists VO which contains a WatchList with the given URI hash
func GetCustomerWatchListsByURIHash(uriHash string) (*CustomerWatchLists, error) {
	return findOneByQuery(&bson.M{"lists.publicurihash": uriHash})
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
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()
	_, err := collection.UpsertId(cw.BsonId, cw)
	return err
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func insertCustomerWatchLists(cw *CustomerWatchLists) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()
	return collection.Insert(cw)
}

func findOne(customerID, sessionID, email string) (*CustomerWatchLists, error) {
	if customerID == "" && sessionID == "" && email == "" {
		return nil, errors.New("Error: Either customerID, sessionID or email of Customer must be provided.")
	}
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	find := &bson.M{}
	if customerID != "" {
		find = &bson.M{"customerID": customerID}
	}
	if sessionID != "" {
		find = &bson.M{"sessionID": sessionID}
	}
	if email != "" {
		find = &bson.M{"email": email}
	}

	customerWatchLists := &CustomerWatchLists{}
	err := collection.Find(find).One(customerWatchLists)
	if err != nil {
		return nil, err
	}

	return customerWatchLists, nil
}
func findOneByQuery(query *bson.M) (*CustomerWatchLists, error) {
	if query == nil {
		return nil, errors.New("Query must not be empty!")
	}
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	customerWatchLists := &CustomerWatchLists{}
	err := collection.Find(query).One(customerWatchLists)
	if err != nil {
		return nil, err
	}

	return customerWatchLists, nil
}
