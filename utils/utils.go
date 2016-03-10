package utils

import (
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/order"
)

// Drops order collection and event_log collection
func DropAllCollections() {
	err := event_log.GetEventPersistor().GetCollection().DropCollection()
	if err != nil {
		panic(err)
	}
	err = order.GetOrderPersistor().GetCollection().DropCollection()
	if err != nil {
		panic(err)
	}
	order.LAST_ASSIGNED_ID = -1
}
