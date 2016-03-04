package shop

import (
	"errors"
	"log"

	"github.com/foomo/shop/order"

	"gopkg.in/mgo.v2/bson"
)

func GetPersistor(db string, collection string) *order.Persistor {
	p, err := order.NewPersistor(db, collection)
	if err != nil {
		panic(err)
	}
	return p
}

func GetShopOrder(db string, collection string, orderID string, customOrderProvider order.OrderCustomProvider) (*order.Order, error) {
	p := GetPersistor(db, collection)
	iter, err := p.Find(&bson.M{"orderid": orderID}, customOrderProvider)
	if err != nil {
		panic(err)
	}
	order, err := iter()
	if err != nil {
		panic(err)
	}
	if order == nil {
		log.Println("Could not find order ", orderID)
		return nil, errors.New("Could not find order with id: " + orderID)
	}
	return order, nil
}
