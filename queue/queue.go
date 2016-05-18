// Package process handles the processing of orders as they change their status
package queue

import (
	"log"

	"github.com/foomo/shop/order"
	"github.com/foomo/shop/persistence"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC TYPES
+++++++++++++++++++++++++++++++++++++++++++++++++ */

type Queue struct {
	persistor      *persistence.Persistor
	processors     []order.Processor
	bulkProcessors []order.BulkProcessor
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC METHODS
+++++++++++++++++++++++++++++++++++++++++++++++++ */

func NewQueue(mongoURL string) (q *Queue, err error) {
	log.Println("NewQueue()...")
	return &Queue{
		persistor: order.GetOrderPersistor(),
	}, nil
}

func (q *Queue) AddProcessor(processor order.Processor) {
	q.processors = append(q.processors, processor)
}

func (q *Queue) AddBulkProcessor(processor order.BulkProcessor) {
	q.bulkProcessors = append(q.bulkProcessors, processor)
}

func (q *Queue) RunProcessor(processor order.Processor) error {
	chanDone := make(chan int)
	chanOrder := make(chan *order.Order)
	go func() {
		i := 0
		running := 0
		done := false
		chanDoneProcessing := make(chan int)
		var waitingOrder *order.Order
		process := func(o *order.Order) {
			running++
			go func() {
				processor.Process(o)
				chanDoneProcessing <- 1
			}()
			//	log.Println("yeah, let us do this concurrently", o.ID, running)

		}

		for !done || running > 0 {
			select {
			case o := <-chanOrder:
				if running < processor.Concurrency() {
					process(o)
					chanOrder <- nil

				} else {
					waitingOrder = o
					//			log.Println("sorry you have to wait")
				}
			case <-chanDoneProcessing:
				i++
				running--
				if waitingOrder != nil {
					process(waitingOrder)
					waitingOrder = nil
					chanOrder <- nil
				}
			case <-chanDone:
				done = true
			}
		}
		//log.Println("exiting with", running, i)
		chanDone <- 1
	}()

	iter, err := order.Find(processor.GetQuery(), processor.OrderCustomProvider())
	if err != nil {
		return err
	}

	for {
		order, err := iter()
		if err != nil {
			log.Println("could not get order", err)
		}
		if order != nil {
			// send to concurrent processing
			chanOrder <- order
			// wait unit we are done
			<-chanOrder
		} else {
			break
		}

	}
	log.Println("done feeding order")
	chanDone <- 1
	<-chanDone
	return nil
}
