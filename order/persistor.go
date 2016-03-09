package order

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/foomo/shop/event_log"
	"github.com/mitchellh/mapstructure"

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

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *Persistor, err error) {
	if GLOBAL_PERSISTOR != nil {
		return GLOBAL_PERSISTOR, nil
	}
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
	GLOBAL_PERSISTOR = &Persistor{
		session:        session,
		db:             parsedURL.Path[1:],
		CollectionName: collectionName,
	}
	return GLOBAL_PERSISTOR, nil
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

func (p *Persistor) InsertOrder(o *Order) error {
	err := p.GetCollection().Insert(o)
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
