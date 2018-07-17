package order

import (
	"errors"
	"log"
	"strconv"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/persistence"
	"github.com/foomo/shop/version"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var (
	globalOrderPersistor         *persistence.Persistor
	globalOrderVersionsPersistor *persistence.Persistor

	orderEnsuredIndexes = []mgo.Index{
		mgo.Index{
			Name:       "id",
			Key:        []string{"id"},
			Unique:     false,
			Background: true,
		},
	}
)

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// GetOrderPersistor will return a singleton instance of an order mongo persistor
func GetOrderPersistor() *persistence.Persistor {
	url := configuration.GetMongoURL()
	collection := configuration.MONGO_COLLECTION_ORDERS
	if globalOrderPersistor == nil {
		p, err := persistence.NewPersistorWithIndexes(url, collection, orderEnsuredIndexes)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB order persistor: " + err.Error()))
		}
		globalOrderPersistor = p
		return globalOrderPersistor
	}

	if url == globalOrderPersistor.GetURL() && collection == globalOrderPersistor.GetCollectionName() {
		return globalOrderPersistor
	}

	p, err := persistence.NewPersistorWithIndexes(url, collection, orderEnsuredIndexes)
	if err != nil || p == nil {
		panic(err)
	}
	globalOrderPersistor = p
	return globalOrderPersistor
}

// GetOrderVersionsPersistor will return a singleton instance of a versioned order mongo persistor
func GetOrderVersionsPersistor() *persistence.Persistor {
	url := configuration.GetMongoURL()
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
			bson.M{"state.key": bson.M{"$ne": OrderStatusInvalid}},
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
