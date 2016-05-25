package examples_test

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/mock"
	"github.com/foomo/shop/order"
	"github.com/foomo/shop/test_utils"
)

func TestSmurfProcessor(t *testing.T) {
	//log.Println("runtime.GOMAXPROCS(16)", runtime.GOMAXPROCS(16))
	test_utils.DropAllCollections()
	q := mock.GetMockQueue()

	const (
		pete = "pete"
		joe  = "joe"
	)

	// add some products in status a

	smurfOrders := map[string]int{
		pete: 1000,
		joe:  2000,
	}

	numberOfOrders := 0
	for smurf, smurfOrderCount := range smurfOrders {
		for i := 0; i < smurfOrderCount; i++ {
			mock.MakeMockOrder(smurf)

			numberOfOrders++
			if numberOfOrders%100 == 0 {
				log.Println(smurf, numberOfOrders)
			}
		}
	}

	log.Println("done writing orders")

	start := time.Now()

	joeProcessor := examples.NewSmurfProcessor(joe)
	peteProcessor := examples.NewSmurfProcessor(pete)

	chanDone := make(chan string)

	a := func() {
		q.RunProcessor(peteProcessor)
		chanDone <- "pete"
	}

	b := func() {
		q.RunProcessor(joeProcessor)
		chanDone <- "joe"
	}
	go a()
	go b()
	log.Println(<-chanDone)
	log.Println(<-chanDone)
	log.Println(time.Now().Sub(start))
	log.Println("done processing")

	fmt.Println("number of orders:", numberOfOrders, ", processed by joe:", joeProcessor.CountProcessed, ", processed by pete:", peteProcessor.CountProcessed)
	// Output: number of orders: 300 , processed by joe: 200 , processed by pete: 100
	if numberOfOrders != smurfOrders["pete"]+smurfOrders["joe"] || joeProcessor.CountProcessed != smurfOrders["joe"] || peteProcessor.CountProcessed != smurfOrders["pete"] {
		t.Fatal("number of orders:", numberOfOrders, ", processed by joe:", joeProcessor.CountProcessed, ", processed by pete:", peteProcessor.CountProcessed)
	}

}

// amount of orders to be generated
const NumOrders = 500

func TestPersistor(t *testing.T) {

	ptp := "Pete the persistor"

	customProvider := examples.SmurfOrderCustomProvider{}
	for i := 0; i < NumOrders; i++ {
		newOrder, err := order.NewOrder(customProvider)
		newOrder.Custom = &examples.SmurfOrderCustom{
			ResponsibleSmurf: ptp,
		}
		err = newOrder.Upsert()
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
	}

	orderIter, err := order.Find(&bson.M{"custom.responsiblesmurf": ptp}, customProvider)
	if err != nil {
		panic(err)
	}
	loadedOrders := []*order.Order{}
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
