package order

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/foomo/shop/persistence"
	"github.com/foomo/shop/version"
	"github.com/mitchellh/mapstructure"

	"github.com/foomo/shop/configuration"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var globalOrderPersistor *persistence.Persistor
var globalOrderVersionsPersistor *persistence.Persistor

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

func Count(query *bson.M, customProvider OrderCustomProvider) (count int, err error) {
	return GetOrderPersistor().GetCollection().Find(query).Count()
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
	// fields := customProvider.Fields()
	// if fields != nil {
	// 	q.Select(fields)
	// }
	_, err = q.Count()
	if err != nil {
		return
	}
	mgoiter := q.Iter()
	iter = func() (o *Order, err error) {
		o = &Order{}

		if mgoiter.Next(o) {
			if customProvider != nil {
				return mapDecode(o, customProvider)
			}
			return o, nil
		}
		return nil, nil
	}
	return
}

func UpsertOrder(o *Order) error {
	// order is unlinked or not yet inserted in db

	if o.unlinkDB || o.BsonId == "" {
		return nil
	}
	p := GetOrderPersistor()

	// Get current version from db and check against verssion of c
	// If they are not identical, there must have been another upsert which would be overwritten by this one.
	// In this case upsert is skipped and an error is returned,
	orderLatestFromDb := &Order{}
	err := p.GetCollection().Find(&bson.M{"id": o.GetID()}).Select(&bson.M{"version": 1}).One(orderLatestFromDb)

	if err != nil {
		log.Println("Upsert failed: Could not find order with id", o.GetID(), "Error:", err)
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
		v := o.Version.Current
		// Set the current version number to keep history consistent
		o.Version.Current = latestVersionInDb
		o.Version.Increment()
		o.Flags.forceUpsert = false
		// Overwrite NumberPrevious, to remember where we came from
		o.Version.Previous = v
	} else {
		o.Version.Increment()
	}

	o.State.SetModified()
	_, err = p.GetCollection().UpsertId(o.BsonId, o)
	if err != nil {
		return err
	}

	// Store version in history
	bsonId := o.BsonId
	o.BsonId = "" // Temporarily reset Mongo ObjectId, so that we can perfrom an Insert.
	pHistory := GetOrderVersionsPersistor()
	pHistory.GetCollection().Insert(o)
	o.BsonId = bsonId // restore bsonId
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
	err := GetOrderPersistor().GetCollection().Remove(bson.M{"_id": o.BsonId})

	return err
}
func DeleteOrderById(id string) error {
	err := GetOrderPersistor().GetCollection().Remove(bson.M{"id": id})
	return err
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetOrderPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_ORDERS
	if globalOrderPersistor == nil {
		p, err := persistence.NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB order persistor: " + err.Error()))
		}
		globalOrderPersistor = p
		return globalOrderPersistor
	}

	if url == globalOrderPersistor.GetURL() && collection == globalOrderPersistor.GetCollectionName() {
		return globalOrderPersistor
	}

	p, err := persistence.NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalOrderPersistor = p
	return globalOrderPersistor
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetOrderVersionsPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_ORDERS_HISTORY
	if globalOrderVersionsPersistor == nil {
		p, err := persistence.NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB order persistor: " + err.Error()))
		}
		globalOrderVersionsPersistor = p
		return globalOrderVersionsPersistor
	}

	if url == globalOrderVersionsPersistor.GetURL() && collection == globalOrderVersionsPersistor.GetCollectionName() {
		return globalOrderVersionsPersistor
	}

	p, err := persistence.NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalOrderVersionsPersistor = p
	return globalOrderVersionsPersistor
}

// GetOrderById returns the order with id
func GetOrderById(id string, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(&bson.M{"id": id}, nil, "", customProvider, false)
}

func GetOrderByQuery(query *bson.M, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(query, nil, "", customProvider, false)
}

// HasCart returns true if an order with state OrderStatusCart exist for customer
func HasCart(customerId string) bool {
	order, _ := GetCart(customerId, nil)
	if order != nil {
		return true
	}
	return false
}

// Get Order for customer which is in state OrderStatusCart
func GetCart(customerId string, customProvider OrderCustomProvider) (*Order, error) {
	query := &bson.M{"customerid:": customerId, "state.key": OrderStatusCart}
	return GetOrderByQuery(query, customProvider)
}
func GetCartID(customerId string) (string, error) {
	query := &bson.M{"customerid:": customerId, "state.key": OrderStatusCart}
	order, err := GetOrderByQuery(query, nil)
	if err != nil {
		return "", err
	}
	return order.GetID(), nil
}

func GetOrdersOfCustomer(customerId string, customProvider OrderCustomProvider) ([]*Order, error) {

	if customProvider == nil {
		return nil, errors.New("Error: customProvider must not be nil")
	}
	// Query for all orders which are neither in OrderStatusCart nor in OrderStatusTechnical
	query := &bson.M{

		"$and": []interface{}{
			bson.M{"customerdata.customerid": customerId},
			bson.M{"state.key": bson.M{"$ne": OrderStatusTechnical}},
			bson.M{"state.key": bson.M{"$ne": OrderStatusCart}},
			bson.M{"state.key": bson.M{"$ne": OrderStatusInvalid}},
		},
	}
	orderIter, err := Find(query, customProvider)
	if err != nil {
		log.Println("Query customerdata.customerid failed", customerId)
		return nil, err
	}
	ordersTmp := []*Order{}
	for {
		o, err := orderIter()
		if err != nil {
			return nil, err
		}
		if o != nil {
			ordersTmp = append(ordersTmp, o)
		} else {
			break
		}
	}
	// reverse order of orders
	orders := []*Order{}
	for i := len(ordersTmp) - 1; i >= 0; i-- {
		orders = append(orders, ordersTmp[i])
	}

	return orders, nil
}

// GetOrderIdsOfCustomer returns all orderIds associated with this customer
func GetOrderIdsOfCustomer(customerId string) ([]string, error) {
	// Query for all orders which are neither in OrderStatusCart nor in OrderStatusTechnical
	query := &bson.M{

		"$and": []interface{}{
			bson.M{"customerdata.customerid": customerId},
			bson.M{"state.key": bson.M{"$ne": OrderStatusTechnical}},
			bson.M{"state.key": bson.M{"$ne": OrderStatusCart}},
		},
	}
	orderIter, err := Find(query, nil) // @TODO this could use a select as we only want the id's
	if err != nil {
		log.Println("Query customerdata.customerid failed:", customerId)
		return nil, err
	}
	idsTmp := []string{}
	for {
		o, err := orderIter()
		if err != nil {
			return nil, err
		}
		if o != nil {
			idsTmp = append(idsTmp, o.GetID())
		} else {
			break
		}
	}

	// reverse order of ids
	ids := []string{}
	for i := len(idsTmp) - 1; i >= 0; i-- {
		ids = append(ids, idsTmp[i])
	}
	return ids, nil

}

func GetCurrentOrderByIdFromVersionsHistory(orderId string, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(&bson.M{"id": orderId}, nil, "-version.current", customProvider, true)
}
func GetCurrentVersionOfOrderFromVersionsHistory(orderId string) (*version.Version, error) {
	order, err := findOneOrder(&bson.M{"id": orderId}, &bson.M{"version": 1}, "-version.current", nil, true)
	if err != nil {
		return nil, err
	}
	return order.GetVersion(), nil
}
func GetOrderByVersion(orderId string, version int, customProvider OrderCustomProvider) (*Order, error) {
	return findOneOrder(&bson.M{"id": orderId, "version.current": version}, nil, "", customProvider, true)
}

func Rollback(orderId string, version int) error {
	currentOrder, err := GetOrderById(orderId, nil)
	if err != nil {
		return err
	}
	if version >= currentOrder.GetVersion().Current || version < 0 {
		return errors.New("Cannot perform rollback to " + strconv.Itoa(version) + " from version " + strconv.Itoa(currentOrder.GetVersion().Current))
	}
	orderFromVersionsHistory, err := GetOrderByVersion(orderId, version, nil)
	if err != nil {
		return err
	}
	// Set bsonId from current order to order from history to overwrite current order on next upsert.
	orderFromVersionsHistory.BsonId = currentOrder.BsonId
	orderFromVersionsHistory.Flags.forceUpsert = true
	return orderFromVersionsHistory.Upsert()

}

func DropAllOrders() error {
	return GetOrderPersistor().GetCollection().DropCollection()

}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// findOneOrder returns one Order from the order database or from the order history database
func findOneOrder(find *bson.M, selection *bson.M, sort string, customProvider OrderCustomProvider, fromHistory bool) (*Order, error) {
	var p *persistence.Persistor
	if fromHistory {
		p = GetOrderVersionsPersistor()
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
	pHistory := GetOrderVersionsPersistor()
	err = pHistory.GetCollection().Insert(o)
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
