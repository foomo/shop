package order_test

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/order"
	"gopkg.in/mgo.v2/bson"
)

func TestPersistor(t *testing.T) {
	p := examples.GetMockPersistor()
	orderCustom := &examples.OrderCustom{
		ResponsibleSmurf: "Pete",
	}
	newOrder := order.NewOrder(orderCustom)
	err := p.Create(newOrder)
	if err != nil {
		panic(err)
	}
	customProvider := examples.FullOrderCustomProvider{}
	loadedOrders, err := p.Find(&bson.M{}, customProvider)
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
