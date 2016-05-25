package order

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/history"
	"github.com/foomo/shop/persistence"
	"github.com/mitchellh/mapstructure"

	"github.com/foomo/shop/configuration"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var globalOrderPersistor *persistence.Persistor
var globalOrderHistoryPersistor *persistence.Persistor

//------------------------------------------------------------------
// ~ CONTRUCTOR
//------------------------------------------------------------------
// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *persistence.Persistor, err error) {
	return persistence.NewPersistor(mongoURL, collectionName)
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// AlreadyExistsInDB checks if a order with given orderID already exists in the database
func AlreadyExistsInDB(orderID string) (bool, error) {
	p := GetOrderPersistor()
	q := p.GetCollection().Find(&bson.M{"id": orderID})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Find returns an iterator for the entries matching on query
func Find(query *bson.M, customProvider OrderCustomProvider) (iter func() (o *Order, err error), err error) {
	p := GetOrderPersistor()
	_, err = p.GetCollection().Find(query).Count()
	if err != nil {
		log.Println(err)
	}
	//log.Println("Persistor.Find(): ", n, "items found for query ", query)
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
	iter = func() (o *Order, err error) {
		o = &Order{}

		if mgoiter.Next(o) {
			return mapDecode(o, customProvider)
		}
		return nil, nil
	}
	return
}

func UpsertOrder(o *Order) error {
	//log.Println("WhoCalledMe: ", utils.WhoCalledMe())
	//log.Println("UPSERT CUSTOMER with id", o.GetID())
	// order is unlinked or not yet inserted in db
	if o.unlinkDB || o.BsonID == "" {
		return nil
	}
	p := GetOrderPersistor()

	// Get current version from db and check against verssion of c
	// If they are not identical, there must have been another upsert which would be overwritten by this one.
	// In this case upsert is skipped and an error is returned,
	orderLatestFromDb := &Order{}
	err := p.GetCollection().Find(&bson.M{"id": o.GetID()}).Select(&bson.M{"version": 1}).One(orderLatestFromDb)

	if err != nil {
		log.Println("Error Upsert Order", err)
		return err
	}

	latestVersionInDb := orderLatestFromDb.Version.GetVersion()
	if latestVersionInDb != o.Version.GetVersion() && !o.Flags.forceUpsert {
		errMsg := fmt.Sprintln("WARNING: Cannot upsert latest version ", strconv.Itoa(latestVersionInDb), "in db with version", strconv.Itoa(o.Version.GetVersion()), "!")
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	if o.Flags.forceUpsert {
		// Remember this number, so that we later know from which version we came from
		v := o.Version.Number
		// Set the current version number to keep history consistent
		o.Version.Number = latestVersionInDb
		o.Version.Increment()
		o.Flags.forceUpsert = false
		// Overwrite NumberPrevious, to remember where we came from
		o.Version.NumberPrevious = v
	} else {
		o.Version.Increment()
	}

	_, err = p.GetCollection().UpsertId(o.BsonID, o)
	if err != nil {
		return err
	}

	// Store version in history
	bsonId := o.BsonID
	o.BsonID = "" // Temporarily reset Mongo ObjectId, so that we can perfrom an Insert.
	pHistory := GetOrderHistoryPersistor()
	pHistory.GetCollection().Insert(o)
	o.BsonID = bsonId // restore bsonId
	event_log.SaveShopEvent(event_log.ActionUpsertingOrder, o.GetID(), err, "")
	return err
}

func UpsertAndGetOrder(o *Order, customProvider OrderCustomProvider) (*Order, error) {
	err := UpsertOrder(o)
	if err != nil {
		return nil, err
	}
	return GetOrderById(o.GetID(), customProvider)
}

func DeleteOrder(o *Order) error {
	err := GetOrderPersistor().GetCollection().Remove(bson.M{"_id": o.BsonID})
	event_log.SaveShopEvent(event_log.ActionDeleteOrder, o.GetID(), err, "")
	return err
}
func DeleteOrderById(id string) error {
	err := GetOrderPersistor().GetCollection().Remove(bson.M{"id": id})
	event_log.SaveShopEvent(event_log.ActionDeleteOrder, id, err, "")
	return err
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetOrderPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_ORDERS
	if globalOrderPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB order persistor: " + err.Error()))
		}
		globalOrderPersistor = p
		return globalOrderPersistor
	}

	if url == globalOrderPersistor.GetURL() && collection == globalOrderPersistor.GetCollectionName() {
		return globalOrderPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalOrderPersistor = p
	return globalOrderPersistor
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetOrderHistoryPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_ORDERS_HISTORY
	if globalOrderHistoryPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB order persistor: " + err.Error()))
		}
		globalOrderHistoryPersistor = p
		return globalOrderHistoryPersistor
	}

	if url == globalOrderHistoryPersistor.GetURL() && collection == globalOrderHistoryPersistor.GetCollectionName() {
		return globalOrderHistoryPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalOrderHistoryPersistor = p
	return globalOrderHistoryPersistor
}

// GetOrderById returns the order with id
func GetOrderById(id string, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(&bson.M{"id": id}, nil, "", customProvider, false)
}

func GetCurrentOrderByIdFromHistory(orderId string, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(&bson.M{"id": orderId}, nil, "-version.number", customProvider, true)
}
func GetCurrentVersionOfOrderFromHistory(orderId string) (*history.Version, error) {
	order, err := findOneOrder(&bson.M{"id": orderId}, &bson.M{"version": 1}, "-version.number", nil, true)
	if err != nil {
		return nil, err
	}
	return order.GetVersion(), nil
}
func GetOrderByVersion(orderId string, version int, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(&bson.M{"id": orderId, "version.number": version}, nil, "", customProvider, true)
}

func Rollback(orderId string, version int) error {
	currentOrder, err := GetOrderById(orderId, nil)
	if err != nil {
		return err
	}
	orderFromHistory, err := GetOrderByVersion(orderId, version, nil)
	if err != nil {
		return err
	}
	// Set bsonId from current order to order from history to overwrite current order on next upsert.
	orderFromHistory.BsonID = currentOrder.BsonID
	orderFromHistory.Flags.forceUpsert = true
	return orderFromHistory.Upsert()

}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// findOneOrder returns one Order from the order database or from the order history database
func findOneOrder(find *bson.M, selection *bson.M, sort string, customProvider OrderCustomProvider, fromHistory bool) (*Order, error) {
	var p *persistence.Persistor
	if fromHistory {
		p = GetOrderHistoryPersistor()
	} else {
		p = GetOrderPersistor()
	}
	order := &Order{}
	if find == nil {
		find = &bson.M{}
	}
	if selection == nil {
		selection = &bson.M{}
	}
	if sort != "" {
		err := p.GetCollection().Find(find).Select(selection).Sort(sort).One(order)
		if err != nil {
			return nil, err
		}
	} else {
		err := p.GetCollection().Find(find).Select(selection).One(order)
		if err != nil {
			return nil, err
		}
	}
	if customProvider != nil {
		var err error
		order, err = mapDecode(order, customProvider)
		if err != nil {
			return nil, err
		}
	}
	event_log.SaveShopEvent(event_log.ActionRetrieveOrder, order.GetID(), nil, "")
	return order, nil
}

// insertOrder inserts a order into the database
func insertOrder(o *Order) error {
	p := GetOrderPersistor()
	alreadyExists, err := AlreadyExistsInDB(o.GetID())
	if err != nil {
		return err
	}
	if alreadyExists {
		log.Println("User with id", o.GetID(), "already exists in the database!")
		return nil
	}
	err = p.GetCollection().Insert(o)
	if err != nil {
		return err
	}
	pHistory := GetOrderHistoryPersistor()
	err = pHistory.GetCollection().Insert(o)
	event_log.SaveShopEvent(event_log.ActionCreateOrder, o.GetID(), err, "")
	return err
}

func mapDecode(o *Order, customProvider OrderCustomProvider) (order *Order, err error) {
	/* Map OrderCustom */
	orderCustom := customProvider.NewOrderCustom()
	if orderCustom != nil && o.Custom != nil {
		err = mapstructure.Decode(o.Custom, orderCustom)
		if err != nil {
			return nil, err
		}
		o.Custom = orderCustom
	}

	/* Map PostionCustom */
	for _, position := range o.Positions {
		positionCustom := customProvider.NewPositionCustom()
		if positionCustom != nil && position.Custom != nil {

			err = mapstructure.Decode(position.Custom, positionCustom)
			if err != nil {
				return nil, err
			}
			position.Custom = positionCustom
		}
	}
	return o, nil
}
