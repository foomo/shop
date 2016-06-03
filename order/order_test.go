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
	err = order.SetState(DefaultStateMachine, OrderStatusConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current State:", order.GetState().Key)

	// This one should not work
	err = order.SetState(DefaultStateMachine, OrderStatusCreated)
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
