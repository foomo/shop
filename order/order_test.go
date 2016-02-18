package order

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/foomo/shop/examples"
	"gopkg.in/mgo.v2/bson"
)

func getMockPersistor() *Persistor {
	url := os.Getenv("SHOP_MONGO_TEST_URL")
	if len(url) == 0 {
		panic("please export SHOP_MONGO_TEST_URL=mongodb://127.0.0.1/foomo-shop-orders")
	}
	p, err := NewPersistor(url)
	if err != nil {
		panic(err)
	}
	p.getCollection().DropCollection()
	return p
}

func TestPersistor(t *testing.T) {
	p := getMockPersistor()
	orderCustom := &examples.OrderCustom{
		ResponsibleSmurf: "Pete",
	}
	newOrder := NewOrder(orderCustom)
	err := p.Create(newOrder)
	if err != nil {
		panic(err)
	}
	customProvider := examples.FullOrderCustomProvider{}
	loadedOrders, err := p.Find(&bson.M{}, nil, customProvider)
	if err != nil {
		panic(err)
	}
	if len(loadedOrders) != 1 {
		t.Fatal("wrong number of orders returned")
	}

	loadedOrder := loadedOrders[0].Custom.(*examples.OrderCustom)

	if !reflect.DeepEqual(loadedOrder, newOrder.Custom) {
		dump("newOrder", newOrder)
		dump("loadedOrder", loadedOrder)

		t.Fatal("should have been equal", loadedOrder, newOrder)
	}
	//LoadOrder(query *bson.M{}, customOrder interface{})
}

func dump(label string, v interface{}) {
	log.Println(label, "::")
	b, _ := json.MarshalIndent(v, "", "   ")
	fmt.Println(reflect.ValueOf(v).Type(), string(b))
}
