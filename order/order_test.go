package order

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/examples"
	"gopkg.in/mgo.v2/bson"
)

// amount of orders to be generated
const NumOrders = 500

func TestPersistor(t *testing.T) {
	p := GetPersistor(configuration.MONGO_URL, configuration.MONGO_COLLECTION_ORDERS)
	ptp := "Pete the persistor"

	for i := 0; i < NumOrders; i++ {
		newOrder := NewOrder()
		newOrder.Custom = &examples.SmurfOrderCustom{
			ResponsibleSmurf: ptp,
		}
		err := p.InsertOrder(newOrder)
		if err != nil {
			panic(err)
		}
	}

	customProvider := examples.SmurfOrderCustomProvider{}
	orderIter, err := p.Find(&bson.M{"custom.responsiblesmurf": ptp}, customProvider)
	if err != nil {
		panic(err)
	}
	loadedOrders := []*Order{}
	for {
		loadedOrder, err := orderIter()
		if loadedOrder != nil {
			loadedOrders = append(loadedOrders, loadedOrder)
		} else {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	t.Log("loaded orders")
	for i, loadedOrder := range loadedOrders {
		t.Log(i, loadedOrder.Custom.(*examples.SmurfOrderCustom).ResponsibleSmurf)
	}

	if len(loadedOrders) != NumOrders {
		t.Fatal("wrong number of orders returned", len(loadedOrders))
	}

	for i, newOrder := range loadedOrders {
		loadedOrder := loadedOrders[i].Custom.(*examples.SmurfOrderCustom)
		if !reflect.DeepEqual(loadedOrder, newOrder.Custom) {
			dump("newOrder", newOrder)
			dump("loadedOrder", loadedOrder)

			t.Fatal("should have been equal", loadedOrder, newOrder)
		}
	}
	//LoadOrder(query *bson.M{}, customOrder interface{})
}

func dump(label string, v interface{}) {
	log.Println(label, "::")
	b, _ := json.MarshalIndent(v, "", "   ")
	fmt.Println(reflect.ValueOf(v).Type(), string(b))
}
