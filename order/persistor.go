package order

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/mitchellh/mapstructure"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Persistor persist *Orders
type Persistor struct {
	session             *mgo.Session
	orderCollectionName string
	db                  string
}

// NewPersistor constructor
func NewPersistor(mongoURL string, orderCollectionName string) (p *Persistor, err error) {
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
		session:             session,
		db:                  parsedURL.Path[1:],
		orderCollectionName: orderCollectionName,
	}
	return
}

func (p *Persistor) GetCollection() *mgo.Collection {
	return p.session.DB(p.db).C(p.orderCollectionName)
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
				return nil, errors.New("Error in Persistor.Find(): Order.Customer is nil. Order must have a Customer!")
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

			// /* Map Order.Custom */
			// err = mapCustom(o.Custom, customProvider.NewOrderCustom())
			// if err != nil {
			// 	return nil, err
			// }
			/* Map Customer.Custom */
			// err := mapCustom(o.Customer.Custom, customProvider.NewCustomerCustom())
			// if err != nil {
			// 	return nil, err
			// }
			//
			// //
			/* Map Postion.Custom */
			// for _, position := range o.Positions {
			// 	err := mapCustom(position.Custom, customProvider.NewPositionCustom())
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// }
			//
			// /* Map Address.Custom */
			// for _, address := range o.Addresses {
			// 	err := mapCustom(address.Custom, customProvider.NewAddressCustom())
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// }

			return o, nil
		}
		return nil, nil
	}

	return
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

func (p *Persistor) Insert(o *Order) error {

	err := p.GetCollection().Insert(o)
	return err
}
