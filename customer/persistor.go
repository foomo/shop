package customer

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/history"
	"github.com/foomo/shop/persistence"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/mgo.v2/bson"
)

// !! NOTE: customer must not import order !!

const (
	VERBOSE = false
)

var globalCustomerPersistor *persistence.Persistor
var globalCustomerHistoryPersistor *persistence.Persistor

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *persistence.Persistor, err error) {
	return persistence.NewPersistor(mongoURL, collectionName)
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetCustomerPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_CUSTOMERS
	if globalCustomerPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB global persistor: " + err.Error()))
		}
		globalCustomerPersistor = p
		return globalCustomerPersistor
	}

	if url == globalCustomerPersistor.GetURL() && collection == globalCustomerPersistor.GetCollectionName() {
		return globalCustomerPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalCustomerPersistor = p
	return globalCustomerPersistor
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetCustomerHistoryPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_CUSTOMERS_HISTORY
	if globalCustomerHistoryPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB order persistor: " + err.Error()))
		}
		globalCustomerHistoryPersistor = p
		return globalCustomerHistoryPersistor
	}

	if url == globalCustomerHistoryPersistor.GetURL() && collection == globalCustomerHistoryPersistor.GetCollectionName() {
		return globalCustomerHistoryPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalCustomerHistoryPersistor = p
	return globalCustomerHistoryPersistor
}

// AlreadyExistsInDB checks if a customer with given customerID already exists in the database
func AlreadyExistsInDB(customerID string) (bool, error) {
	p := GetCustomerPersistor()
	q := p.GetCollection().Find(&bson.M{"id": customerID})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindOne returns one single customer
func GetCustomerById(id string, customProvider CustomerCustomProvider) (*Customer, error) {
	p := GetCustomerPersistor()
	customer := &Customer{}
	err := p.GetCollection().Find(&bson.M{"id": id}).One(customer)
	if err != nil {
		return nil, err
	}
	customer, err = mapDecode(customer, customProvider)
	event_log.SaveShopEvent(event_log.ActionRetrieveCustomer, id, err, "")
	return customer, err
}

// Find returns an iterator for all entries found matching on query.
func Find(query *bson.M, customProvider CustomerCustomProvider) (iter func() (cust *Customer, err error), err error) {
	p := GetCustomerPersistor()
	_, err = p.GetCollection().Find(query).Count()
	if err != nil {
		log.Println(err)
	}
	q := p.GetCollection().Find(query)
	fields := customProvider.Fields()
	if fields != nil {
		q.Select(fields)
	}
	_, err = q.Count()
	if err != nil {
		return
	}
	mgoiter := q.Iter()
	iter = func() (cust *Customer, err error) {
		cust = &Customer{}
		if mgoiter.Next(cust) {
			return mapDecode(cust, customProvider)
		}
		return nil, nil
	}
	return
}

func mapDecode(cust *Customer, customProvider CustomerCustomProvider) (customer *Customer, err error) {
	/* Map CustomerCustom */
	customerCustom := customProvider.NewCustomerCustom()
	if customerCustom != nil && cust.Custom != nil {
		err = mapstructure.Decode(cust.Custom, customerCustom)
		if err != nil {
			return nil, err
		}
		cust.Custom = customerCustom
	}

	/* Map AddressCustom */
	for _, address := range cust.Addresses {
		addressCustom := customProvider.NewAddressCustom()
		if addressCustom != nil && address.Custom != nil {

			err = mapstructure.Decode(address.Custom, addressCustom)
			if err != nil {
				return nil, err
			}
			address.Custom = addressCustom
		}
	}
	return cust, nil
}

func InsertCustomer(c *Customer) error {
	p := GetCustomerPersistor()
	alreadyExists, err := AlreadyExistsInDB(c.GetID())
	if err != nil {
		return err
	}
	if alreadyExists {
		log.Println("User with id", c.GetID(), "already exists in the database!")
		return nil
	}
	err = p.GetCollection().Insert(c)
	if err != nil {
		return err
	}
	pHistory := GetCustomerHistoryPersistor()
	err = pHistory.GetCollection().Insert(c)
	event_log.SaveShopEvent(event_log.ActionCreateCustomer, c.GetID(), err, "")
	return err
}

func UpsertCustomer(c *Customer) error {
	//log.Println("UPSERT CUSTOMER with id", c.GetID())
	// order is unlinked or not yet inserted in db
	if c.unlinkDB || c.BsonID == "" {
		return nil
	}
	p := GetCustomerPersistor()

	// Get current version from db and check against verssion of c
	// If they are not identical, there must have been another upsert which would be overwritten by this one.
	// In this case upsert is skipped and an error is returned,
	customerLatestFromDb := &Customer{}
	err := p.GetCollection().Find(&bson.M{"id": c.GetID()}).Select(&bson.M{"version": 1}).One(customerLatestFromDb)

	if err != nil {
		log.Println("ERROR", err)
		return err
	}

	latestVersionInDb := customerLatestFromDb.Version.GetVersion()
	if latestVersionInDb != c.Version.GetVersion() {
		errMsg := fmt.Sprintln("WARNING: Cannot upsert latest version ", strconv.Itoa(latestVersionInDb), "in db with version", strconv.Itoa(c.Version.GetVersion()), "!")
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	c.Version.Number = latestVersionInDb
	c.Version.Increment()

	_, err = p.GetCollection().UpsertId(c.BsonID, c)
	if err != nil {
		return err
	}

	// Store version in history
	bsonId := c.BsonID
	c.BsonID = "" // Temporarily reset Mongo ObjectId, so that we can perfrom an Insert.
	pHistory := GetCustomerHistoryPersistor()
	pHistory.GetCollection().Insert(c)
	c.BsonID = bsonId // restore bsonId
	event_log.SaveShopEvent(event_log.ActionUpsertingCustomer, c.GetID(), err, "")
	return err
}

func UpsertAndGetCustomer(c *Customer, customProvider CustomerCustomProvider) (*Customer, error) {
	err := UpsertCustomer(c)
	if err != nil {
		return nil, err
	}
	return GetCustomerById(c.GetID(), customProvider)
}

func DeleteCustomer(c *Customer) error {
	err := GetCustomerPersistor().GetCollection().Remove(bson.M{"_id": c.BsonID})
	event_log.SaveShopEvent(event_log.ActionDeleteCustomer, c.GetID(), err, "")
	return err
}
func DeleteCustomerById(id string) error {
	err := GetCustomerPersistor().GetCollection().Remove(bson.M{"id": id})
	event_log.SaveShopEvent(event_log.ActionDeleteCustomer, id, err, "")
	return err
}
func GetCurrentCustomerFromHistory(customProvider CustomerCustomProvider) (*Customer, error) {
	customer := &Customer{}
	p := GetCustomerHistoryPersistor()
	err := p.GetCollection().Find(&bson.M{}).Sort("-version.number").One(customer)
	if err != nil {
		return nil, err
	}
	return mapDecode(customer, customProvider)
}
func GetCurrentVersionFromHistory() (*history.Version, error) {
	customer := &Customer{}
	p := GetCustomerHistoryPersistor()
	err := p.GetCollection().Find(&bson.M{}).Select(&bson.M{"version": 1}).Sort("-version.number").One(customer)
	if err != nil {
		return nil, err
	}
	log.Println("Number", customer.Version.Number, "Time", customer.Version.TimeStamp)
	return customer.GetVersion(), nil
}
func GetCustomerByVersion(version int, customProvider CustomerCustomProvider) (*Customer, error) {
	customer := &Customer{}
	p := GetCustomerHistoryPersistor()
	err := p.GetCollection().Find(&bson.M{"version.number": version}).One(customer)
	if err != nil {
		return nil, err
	}
	return mapDecode(customer, customProvider)
}
