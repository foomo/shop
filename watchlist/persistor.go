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

const (
	KeyAddrkey   = "addrkey"
	KeySessionID = "sessionID"
	KeyEmail     = "email"
)

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

func NewCustomerWatchListsFromAddrKey(addrKey string) (*CustomerWatchLists, error) {
	exists, err := CustomerWatchListsExists(addrKey, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("CustomerWatchLists for addrKey " + addrKey + " already exists.")
	}
	err = insertCustomerWatchLists(&CustomerWatchLists{
		AddrKey: addrKey,
		Lists:   []*WatchList{},
	})
	if err != nil {
		return nil, err
	}
	return GetCustomerWatchListsByAddrKey(addrKey)
}
func NewCustomerWatchListsFromSessionID(sessionID string) (*CustomerWatchLists, error) {
	exists, err := CustomerWatchListsExists("", sessionID)
	if err != nil {
		return nil, err
	}
	if exists {
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

func CustomerWatchListsExists(addrKey, sessionId string) (bool, error) {

	key := ""
	value := ""
	if addrKey != "" {
		key = KeyAddrkey
		value = addrKey
	} else if sessionId != "" {
		key = KeySessionID
		value = sessionId
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
func DeleteCustomerWatchListsByAddrKey(addrKey string) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	return collection.Remove(bson.M{KeyAddrkey: addrKey})
}
func DeleteCustomerWatchListsBySessionId(id string) error {
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	return collection.Remove(bson.M{KeySessionID: id})
}

// GetCustomerWatchListsByURIHash returns the CustomerWatchLists VO which contains a WatchList with the given URI hash
func GetCustomerWatchListsByURIHash(uriHash string) (*CustomerWatchLists, error) {
	return findOneByQuery(&bson.M{"lists.publicurihash": uriHash})
}
func GetCustomerWatchListsByAddrKey(addrKey string) (*CustomerWatchLists, error) {
	return findOne(addrKey, "")
}
func GetCustomerWatchListsBySessionID(sessionID string) (*CustomerWatchLists, error) {
	return findOne("", sessionID)
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

func findOne(addrKey, sessionID string) (*CustomerWatchLists, error) {
	if addrKey == "" && sessionID == "" {
		return nil, errors.New("Either addrKey or sessionID be provided.")
	}
	session, collection := GetWatchListPersistor().GetCollection()
	defer session.Close()

	find := &bson.M{}
	if addrKey != "" {
		find = &bson.M{KeyAddrkey: addrKey}
	}
	if sessionID != "" {
		find = &bson.M{KeySessionID: sessionID}
	}

	customerWatchLists := &CustomerWatchLists{}
	err := collection.Find(find).Sort("-_id").One(customerWatchLists)
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
	err := collection.Find(query).Sort("-_id").One(customerWatchLists)
	if err != nil {
		return nil, err
	}

	return customerWatchLists, nil
}
