package order

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Persistor persist *Orders
type Persistor struct {
	session *mgo.Session
	db      string
}

// NewPersistor constructor
func NewPersistor(mongoURL string) (p *Persistor, err error) {
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
		session: session,
		db:      parsedURL.Path[1:],
	}
	return
}

func (p *Persistor) getCollection() *mgo.Collection {
	return p.session.DB(p.db).C("orders")
}

func (p *Persistor) Find(query *bson.M, customOrderConstructor func() interface{}) (orders []*Order, err error) {
	q := p.getCollection().Find(query)
	_, err = q.Count()
	if err != nil {
		return
	}
	orders = []*Order{}
	iter := q.Iter()
	for {
		order := &Order{}
		if iter.Next(order) {
			customOrder := customOrderConstructor()
			err = mapstructure.Decode(order.Custom, customOrder)
			if err != nil {
				return
			}
			order.Custom = customOrder
			orders = append(orders, order)
		} else {
			break
		}

	}
	return
}

func (p *Persistor) Create(o *Order) error {
	err := p.getCollection().Insert(o)
	return err
}
