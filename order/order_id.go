package order

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/foomo/shop/configuration"
	"gopkg.in/mgo.v2/bson"
)

// TODO: this should not be part of the general shop

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			CONSTANTS / VARS
+++++++++++++++++++++++++++++++++++++++++++++++++ */
var LAST_ASSIGNED_ID int = -1
var OrderIDLock sync.Mutex

/* ++++++++++++++++++++++++++++++++++++++++++++++++
			PUBLIC TYPES
+++++++++++++++++++++++++++++++++++++++++++++++++ */

type OrderIDWrapper struct {
	OrderID string
}

// createOrderID creates a new order id within specified range in foomo/shop/configuration (ids cycle when range is exceeded).
func createOrderID() (id string, err error) {
	// Globus specifec prefix
	prefix := "000"
	p := GetOrderPersistor()
	OrderIDLock.Lock()

	// Application has been restarted. LAST_ASSIGNED_ID is not yet initialized
	if LAST_ASSIGNED_ID == -1 {
		// Retrieve orderID of the most recent order
		q := p.GetCollection().Find(&bson.M{}).Sort("-_id").Limit(1).Select(&bson.M{"id": true})
		iter := q.Iter()
		c, err := q.Count()
		if err != nil {
			OrderIDLock.Unlock()
			return id, err
		}
		// If no orders exist, start with first value of range
		if c == 0 {
			// Database is emtpy. Use first id from specified id range
			fmt.Println("Database is emtpy. Use first id from specified id range")
			LAST_ASSIGNED_ID = configuration.ORDER_ID_RANGE[0]
			OrderIDLock.Unlock()
			return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil // "000" prefix is custom for Globus
		}
		orderIDWrapper := &OrderIDWrapper{}
		iter.Next(orderIDWrapper)
		//log.Println("orderIDWrapper.orderID:", orderIDWrapper.OrderID)
		idInt, err := strconv.Atoi(orderIDWrapper.OrderID)
		if err != nil {
			panic(err)
		}
		LAST_ASSIGNED_ID = idInt + 1
		OrderIDLock.Unlock()
		return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil
	}
	// if range is exceeded, use first value of range again
	if LAST_ASSIGNED_ID == configuration.ORDER_ID_RANGE[1] {
		LAST_ASSIGNED_ID = configuration.ORDER_ID_RANGE[0]
		OrderIDLock.Unlock()
		return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil
	}

	// increment orderID
	LAST_ASSIGNED_ID = LAST_ASSIGNED_ID + 1
	OrderIDLock.Unlock()
	return prefix + strconv.Itoa(LAST_ASSIGNED_ID), nil

}
