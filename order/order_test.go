package order

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForceUpsert(t *testing.T) {

	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()

	assert.NoError(t, collection.DropCollection(), "clean up")

	orderID := "Foo"
	fnOrderID := func() (string, error) {
		return orderID, nil
	}

	fnLog := func(t *testing.T, order *Order, comment string) {
		t.Log()
		t.Log(comment)
		t.Log("id:", order.Id, "site:", order.Site)
		t.Log("version previous:", order.Version.Previous, "current:", order.Version.Current)
	}

	// create order
	order, err := NewOrderWithCustomId(nil, fnOrderID)
	assert.NoError(t, err, "NewOrder")
	fnLog(t, order, "new order")
	// retrieve order from db
	order, err = GetOrderById(orderID, nil)
	assert.NoError(t, err, "GetOrderById")
	fnLog(t, order, "load again")

	// change order
	order.Site = "globus"
	assert.NoError(t, order.Upsert(), "upsert")
	// get changed order
	orderGlobus, err := GetOrderById(orderID, nil)
	assert.NoError(t, err, "GetOrderById")
	fnLog(t, orderGlobus, "load after upsert, to be used for later rollback")
	// check if change is there
	assert.Equal(t, "globus", orderGlobus.Site)

	// change again
	orderNavyBoot, err := GetOrderById(orderID, nil)
	assert.NoError(t, err, "GetOrderById")
	fnLog(t, orderNavyBoot, "load again")
	orderNavyBoot.Site = "navyboot"
	assert.NoError(t, orderNavyBoot.Upsert(), "upsert")

	// load order again
	orderNavyBoot, err = GetOrderById(orderID, nil)
	assert.NoError(t, err, "GetOrderById")
	fnLog(t, orderNavyBoot, "load after upsert")
	// check if change is there
	assert.Equal(t, "navyboot", orderNavyBoot.Site)

	// rollback to orderGlobus
	orderGlobus.SetForceUpsert(true)
	assert.Equal(t, true, orderGlobus.Flags.forceUpsert)
	fnLog(t, orderGlobus, "version of order used for rollback")
	assert.NoError(t, orderGlobus.Upsert(), "upsert")

	// check of site is globus again
	orderGlobus, err = GetOrderById(orderID, nil)
	assert.NoError(t, err, "GetOrderById")
	fnLog(t, orderGlobus, "load after rollback")
	// check if change is there
	assert.Equal(t, "globus", orderGlobus.Site)
	assert.Equal(t, false, orderGlobus.Flags.forceUpsert)

}

// Test transitions between states
func TestOrderStatusTransition(t *testing.T) {
	DropAllOrders()
	order, err := NewOrder(nil)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Current State:", order.GetState().Key)

	// This state transistion should work
	err = order.SetState(DefaultStateMachine, OrderStatusConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should not work
	err = order.SetState(DefaultStateMachine, OrderStatusCart)
	if err != nil {
		log.Println(err)
	} else {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should work
	err = order.ForceState(DefaultStateMachine, OrderStatusComplete)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should work
	err = order.ForceState(DefaultStateMachine, OrderStatusInvalid)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should work
	err = order.ForceState(DefaultStateMachine, OrderStatusComplete)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)
}
