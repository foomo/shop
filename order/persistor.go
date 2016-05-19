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
			/* Map OrderCustom */
			orderCustom := customProvider.NewOrderCustom()
			if orderCustom != nil && o.Custom != nil {
				err = mapstructure.Decode(o.Custom, orderCustom)
				if err != nil {
					return nil, err
				}
				o.Custom = orderCustom
			}

			// /* Map CustomerCustom */
			// customerCustom := customProvider.NewCustomerCustom()
			// if o.Customer == nil {
			// 	return nil, errors.New("Error in Persistor.Find(): Customer is nil. Order must have a Customer!")
			// }
			// if customerCustom != nil && o.Customer.Custom != nil {
			// 	err = mapstructure.Decode(o.Customer.Custom, customerCustom)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	o.Customer.Custom = customerCustom
			// }

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
		return nil, nil
	}

	return
}

/* Retrieve a single position for an order */
// This example query works in mongo shell:
// query: {orderid:"0009200001"}, fields: {positions:{$elemMatch:{"custom.positionnumber":1}}}
// db.order_story.find({orderid:"0009200001"}, {positions:{$elemMatch:{"custom.positionnumber":1}}}).pretty()
// func GetPositionOfOrder(query *bson.M, fields *bson.M, customProvider OrderCustomProvider) (*Position, error) {
// 	p := GetOrderPersistor()
// 	n, err := p.GetCollection().Find(query).Count()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if n != 1 {
// 		return nil, errors.New("Error: Query was not unique!")
// 	}
// 	event_log.Debug("Persistor.Find(): ", n, "items found for query ", query)
// 	q := p.GetCollection().Find(query)
// 	q.Select(fields)
// 	iter := q.Iter()
// 	position := &Position{}
// 	if iter.Next(position) {
// 		// if unmarshalling into postion was successful, set PositionCustom
// 		// TODO this is duplicate code, make separate function
// 		positionCustom := customProvider.NewPositionCustom()
// 		if positionCustom != nil && position.Custom != nil {

// 			err = mapstructure.Decode(position.Custom, positionCustom)
// 			if err != nil {
// 				return nil, err
// 			}
// 			position.Custom = positionCustom
// 		}
// 		return position, nil
// 	}
// 	return nil, errors.New("Could not retrieve position: ")

// }

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
	event_log.SaveShopEvent(event_log.ActionInsertingOrder, o.GetID(), err, "")
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

func InsertEvent(e *event_log.Event) error {
	p := GetOrderPersistor()
	err := p.GetCollection().Insert(e)
	return err
}

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

// GetShopOrder retrieves a single order from the database
func GetShopOrder(orderID string, customOrderProvider OrderCustomProvider) *Order {
	iter, err := Find(&bson.M{"id": orderID}, customOrderProvider)
	if err != nil {
		log.Println(err.Error())
		event_log.SaveShopEvent(event_log.ActionRetrieveOrder, orderID, err, "")
		return nil
	}
	order, err := iter()
	if err != nil {
		log.Println(err.Error())
		event_log.SaveShopEvent(event_log.ActionRetrieveOrder, orderID, err, "")
		return nil
	}
	if order == nil {
		log.Println(err.Error())
		event_log.SaveShopEvent(event_log.ActionRetrieveOrder, orderID, errors.New("Order is nil"), "")
		return nil
	}
	return order
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PRIVATE METHODS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

/* this does not work yet */
// func mapCustom(m interface{}, raw interface{}) error {
// 	if raw == nil {
// 		return errors.New("raw must not be nil")
// 	}
// 	err := mapstructure.Decode(m, raw)
// 	if err != nil {
// 		return err
// 	}
// 	m = raw
// 	return nil
// }
