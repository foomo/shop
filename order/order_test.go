package order

import (
	"log"
	"testing"
)

// Test transitions between states
func TestOrderStatusTransition(t *testing.T) {
	DropAllOrders()
	order, err := NewOrder(nil)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Current State:", order.GetState().Key)

	// This state transistion should work
	err = order.SetState(OrderStatusConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should not work
	err = order.SetState(OrderStatusCreated)
	if err != nil {
		log.Println(err)
	} else {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should work
	err = order.ForceState(OrderStatusComplete)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should work
	err = order.ForceState(OrderStatusInvalid)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should work
	err = order.ForceState(OrderStatusComplete)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)
}
