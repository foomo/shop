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

func (p *Persistor) GetCollection() *mgo.Collection {
	return p.session.DB(p.db).C("orders")
}

func (p *Persistor) Find(query *bson.M, customProvider OrderCustomProvider) (iter func() (o *Order, err error), err error) {
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
			orderCustom := customProvider.NewOrderCustom()
			if orderCustom != nil {
				err = mapstructure.Decode(o.Custom, orderCustom)
				if err != nil {
					return nil, err
				}
				o.Custom = orderCustom
			}
			return o, nil
		}
		// eof ?!
		return nil, nil
	}

	return
}

func (p *Persistor) Insert(o *Order) error {
	err := p.GetCollection().Insert(o)
	return err
}
