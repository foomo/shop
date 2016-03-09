package order

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/foomo/shop/event_log"
	"github.com/mitchellh/mapstructure"

	"strconv"

	"sync"

	"github.com/foomo/shop/configuration"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Persistor persist *Orders
type Persistor struct {
	session        *mgo.Session
	CollectionName string
	db             string
}

var GLOBAL_PERSISTOR *Persistor
var LAST_ASSIGNED_ID int = -1
var OrderIDLock sync.Mutex

type OrderIDWrapper struct {
	OrderID string
}

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *Persistor, err error) {
	parsedURL, err := url.Parse(mongoURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "mongodb" {
		return nil, fmt.Errorf("missing scheme mongo:// in %q", mongoURL)
	}
	if len(parsedURL.Path) < 2 {
		return nil, errors.New("invalid mongoURL missing db should be mongodb://server:port/db")
	}
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		return nil, err
	}
	p = &Persistor{
		session:        session,
		db:             parsedURL.Path[1:],
		CollectionName: collectionName,
	}
	return p, nil
}

func (p *Persistor) GetCollection() *mgo.Collection {
	return p.session.DB(p.db).C(p.CollectionName)
}

func (p *Persistor) Find(query *bson.M, customProvider OrderCustomProvider) (iter func() (o *Order, err error), err error) {
	n, err := p.GetCollection().Find(query).Count()
	if err != nil {
		log.Println(err)
	}
	log.Println("Persistor.Find(): ", n, "items found for query ", query)
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

			/* Map CustomerCustom */
			customerCustom := customProvider.NewCustomerCustom()
			if o.Customer == nil {
				return nil, errors.New("Error in Persistor.Find(): Customer is nil. Order must have a Customer!")
			}
			if customerCustom != nil && o.Customer.Custom != nil {
				err = mapstructure.Decode(o.Customer.Custom, customerCustom)
				if err != nil {
					return nil, err
				}
				o.Customer.Custom = customerCustom
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

			/* Map AddressCustom */
			for _, address := range o.Addresses {
				addressCustom := customProvider.NewAddressCustom()
				if addressCustom != nil && address.Custom != nil {

					err = mapstructure.Decode(address.Custom, addressCustom)
					if err != nil {
						return nil, err
					}
					address.Custom = addressCustom
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
func (p *Persistor) GetPositionOfOrder(query *bson.M, fields *bson.M, customProvider OrderCustomProvider) (*Position, error) {
	n, err := p.GetCollection().Find(query).Count()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, errors.New("Error: Query was not unique!")
	}
	log.Println("Persistor.Find(): ", n, "items found for query ", query)
	q := p.GetCollection().Find(query)
	q.Select(fields)
	iter := q.Iter()
	position := &Position{}
	if iter.Next(position) {
		// if unmarshalling into postion was successful, set PositionCustom
		// TODO this is duplicate code, make separate function
		positionCustom := customProvider.NewPositionCustom()
		if positionCustom != nil && position.Custom != nil {

			err = mapstructure.Decode(position.Custom, positionCustom)
			if err != nil {
				return nil, err
			}
			position.Custom = positionCustom
		}
		return position, nil
	}
	return nil, errors.New("Could not retrieve position: ")

}

/* this does not work yet */
func mapCustom(m interface{}, raw interface{}) error {
	if raw == nil {
		return errors.New("raw must not be nil")
	}
	err := mapstructure.Decode(m, raw)
	if err != nil {
		return err
	}
	m = raw
	return nil
}

/*
Create new orderID within specified range (ids cycle when range is exceeded).
*/
func createOrderID() (id string, err error) {
	// Globus specifec prefix
	prefix := "000"
	p := GetPersistor(configuration.MONGO_URL, configuration.MONGO_COLLECTION_ORDERS)

	OrderIDLock.Lock()

	// Application has been restarted. LAST_ASSIGNED_ID is not yet initialized
	if LAST_ASSIGNED_ID == -1 {
		// Retrieve orderID of the most recent order
		q := p.GetCollection().Find(&bson.M{}).Sort("-_id").Limit(1).Select(&bson.M{"orderid": true})
		iter := q.Iter()
		c, err := q.Count()
		if err != nil {
			OrderIDLock.Unlock()
			return id, err
		}
		// If no orders exist, start with first value of range
		if c == 0 {
			// Database is emtpy. Use first id from specified id range
			fmt.Println("Database is emtpy. Use first id from specified id range")
			LAST_ASSIGNED_ID = configuration.ORDER_ID_RANGE[0]
			OrderIDLock.Unlock()
			return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil // "000" prefix is custom for Globus
		}
		orderIDWrapper := &OrderIDWrapper{}
		iter.Next(orderIDWrapper)
		idInt, err := strconv.Atoi(orderIDWrapper.OrderID)
		if err != nil {
			panic(err)
		}
		LAST_ASSIGNED_ID = idInt + 1
		OrderIDLock.Unlock()
		return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil
	}
	// if range is exceeded, use first value of range again
	if LAST_ASSIGNED_ID == configuration.ORDER_ID_RANGE[1] {
		LAST_ASSIGNED_ID = configuration.ORDER_ID_RANGE[0]
		OrderIDLock.Unlock()
		return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil
	}

	// increment orderID
	LAST_ASSIGNED_ID = LAST_ASSIGNED_ID + 1
	OrderIDLock.Unlock()
	return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil

}

// Create unique OrderID and insert order in database
func (p *Persistor) InsertOrder(o *Order) (err error) {
	o.OrderID, err = createOrderID()
	if err != nil {
		return err
	}
	//log.Println("Created orderID:", o.OrderID)
	err = p.GetCollection().Insert(o)
	event_log.SaveShopEvent(event_log.ActionInsertingOrder, o.OrderID, err)
	return err
}

func (p *Persistor) UpsertOrder(o *Order) (*mgo.ChangeInfo, error) {
	info, err := p.GetCollection().UpsertId(o.ID, o)
	event_log.SaveShopEvent(event_log.ActionUpsertingOrder, o.OrderID, err)
	return info, err
}

func (p *Persistor) InsertEvent(e *event_log.Event) error {
	err := p.GetCollection().Insert(e)
	return err
}

func GetPersistor(db string, collection string) *Persistor {
	if GLOBAL_PERSISTOR == nil {
		p, err := NewPersistor(db, collection)
		if err != nil {
			panic(err)
		}
		return p
	}
	if db == GLOBAL_PERSISTOR.db && collection == GLOBAL_PERSISTOR.CollectionName {
		return GLOBAL_PERSISTOR
	}
	p, err := NewPersistor(db, collection)
	if err != nil {
		panic(err)
	}
	return p
}

func GetShopOrder(db string, collection string, orderID string, customOrderProvider OrderCustomProvider) *Order {
	p := GetPersistor(db, collection)
	iter, err := p.Find(&bson.M{"orderid": orderID}, customOrderProvider)
	if err != nil {
		log.Println("GetShopOrder(): Found no matching order for orderID " + orderID)
		panic(err)
	}
	order, err := iter()
	if err != nil {
		log.Println("GetShopOrder(): Found no matching order for orderID " + orderID)
		panic(err)
	}
	if order == nil {
		log.Println("GetShopOrder(): Found no matching order for orderID " + orderID)
		return nil
	}
	return order
}
