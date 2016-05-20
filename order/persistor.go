package order

import (
	"errors"
	"log"

	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/persistence"
	"github.com/mitchellh/mapstructure"

	"github.com/foomo/shop/configuration"
	"gopkg.in/mgo.v2/bson"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			CONSTANTS / VARS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

var globalOrderPersistor *persistence.Persistor

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC TYPES
+++++++++++++++++++++++++++++++++++++++++++++++++ */

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS ON PERSISTOR
+++++++++++++++++++++++++++++++++++++++++++++++++ */

// FindOne returns one single order
func GetOrderById(id string, customProvider OrderCustomProvider) (*Order, error) {
	p := GetOrderPersistor()
	order := &Order{}
	err := p.GetCollection().Find(&bson.M{"id": id}).One(order)
	if err != nil {
		return nil, err
	}
	order, err = mapDecode(order, customProvider)
	event_log.SaveShopEvent(event_log.ActionRetrieveOrder, id, err, "")
	return order, err
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

// Create (or override) unique OrderID and insert order in database
func InsertOrder(o *Order) (err error) {
	p := GetOrderPersistor()
	newID, err := createOrderID() // TODO This is not Globus Specific and should not be in shop
	if err != nil {
		return err
	}
	o.OverrideId(newID)
	o.SetStatus(OrderStatusCreated)
	//log.Println("Created orderID:", o.OrderID)

	err = p.GetCollection().Insert(o)
	event_log.SaveShopEvent(event_log.ActionCreateOrder, o.GetID(), err, "")
	return err
}

func UpsertOrder(o *Order) error {
	// order is unlinked or not yet inserted in db
	if o.unlinkDB || o.BsonID == "" {
		return nil
	}
	p := GetOrderPersistor()
	_, err := p.GetCollection().UpsertId(o.BsonID, o)
	if err != nil {
		panic(err)
	}
	event_log.SaveShopEvent(event_log.ActionUpsertingOrder, o.GetID(), err, "")
	return err
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

// func InsertEvent(e *event_log.Event) error {
// 	p := GetOrderPersistor()
// 	err := p.GetCollection().Insert(e)
// 	return err
// }

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *persistence.Persistor, err error) {
	return persistence.NewPersistor(mongoURL, collectionName)
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
