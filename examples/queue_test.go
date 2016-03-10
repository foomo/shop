package examples_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/mock"
	"github.com/foomo/shop/order"
	"github.com/foomo/shop/utils"
)

func TestSmurfProcessor(t *testing.T) {
	//log.Println("runtime.GOMAXPROCS(16)", runtime.GOMAXPROCS(16))
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

	p := order.GetOrderPersistor()
	utils.DropAllCollections()

	numberOfOrders := 0
	for smurf, smurfOrderCount := range smurfOrders {
		for i := 0; i < smurfOrderCount; i++ {
			o := mock.MakeMockOrder(smurf)

			p.InsertOrder(o)

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
