package utils

import (
	"log"

	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/order"
)

// Drops order collection and event_log collection
func DropAllCollections() {
	err := event_log.GetEventPersistor().GetCollection().DropCollection()
	if err != nil {
		// Do not panic here. If db does not yet exist, it is ok for DropCollection to fail.
		log.Println("Error: EventPersistor DropCollection() ", err)
	}
	err = order.GetOrderPersistor().GetCollection().DropCollection()
	if err != nil {
		// Do not panic here. If db does not yet exist, it is ok for DropCollection to fail.
		log.Println("Error: EventPersistor DropCollection() ", err)
	}
	order.LAST_ASSIGNED_ID = -1
}
