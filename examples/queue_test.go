package examples_test

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/foomo/shop/examples"
)

func ExampleGarbageProcessor() {
	log.Println("runtime.GOMAXPROCS(16)", runtime.GOMAXPROCS(16))
	q := examples.GetMockQueue()

	const (
		pete = "pete"
		joe  = "joe"
	)

	// add some products in status a

	smurfOrders := map[string]int{
		pete: 100000,
		joe:  100000,
	}

	p := examples.GetMockPersistor()
	numberOfOrders := 0
	for smurf, smurfOrderCount := range smurfOrders {
		for i := 0; i < smurfOrderCount; i++ {
			o := examples.MakeMockOrder(smurf)
			p.Insert(o)
			numberOfOrders++
			if numberOfOrders%1000 == 0 {
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
	go a()
	go b()
	log.Println(<-chanDone)
	log.Println(<-chanDone)

	for i := 0; i < 20; i++ {
		log.Println("sleeping", i)
		time.Sleep(time.Second)
	}

	fmt.Println("number of orders:", numberOfOrders, ", processed by joe:", joeProcessor.CountProcessed, ", processed by pete:", peteProcessor.CountProcessed)
	// Output: number of orders: 13 , processed by joe: 3 , processed by pete: 10

}
