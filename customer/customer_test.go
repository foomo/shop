package customer

import (
	"log"
	"testing"
)

func TestAppLogicGetLatestCustomerFromDb(t *testing.T) {
	DropAllCustomers()
	customer, err := NewCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Perform 3 Upserts
	customer.Person.FirstName = "Foo"
	err = customer.Upsert()
	customer.Person.MiddleName = "Bob"
	err = customer.Upsert()
	customer.Person.LastName = "Bar"
	err = customer.Upsert()
	if err != nil {
		t.Fatal(err)
	}

	//Check if version number is 3
	customer, err = GetCurrentCustomerByIdFromHistory(customer.GetID(), nil)
	if customer.GetVersion().Number != 3 {
		log.Println("Version is ", customer.GetVersion().Number, "- should have been 3.")
		t.Fail()
	}
}
func TestAppLogicDiff2LatestCustomerVersions(t *testing.T) {
	customer1, _ := create2CustomersAndPerformSomeUpserts(t)

	_, err := DiffTwoLatestCustomerVersions(customer1.GetID(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
}
func TestAppLogicRollbackAndDiff(t *testing.T) {
	customer1, _ := create2CustomersAndPerformSomeUpserts(t)

	errRoll := customer1.Rollback(customer1.GetVersion().Number - 1)
	if errRoll != nil {
		t.Fatal(errRoll)
	}
	customer1, errRoll = GetCustomerById(customer1.GetID(), nil)

	_, err := DiffTwoLatestCustomerVersions(customer1.GetID(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
}
func TestAppLogicRollback(t *testing.T) {
	customer1, _ := create2CustomersAndPerformSomeUpserts(t)
	log.Println("Version", customer1.GetVersion(), "FirstName", customer1.Person.FirstName)
	err := customer1.Rollback(customer1.GetVersion().Number - 1)
	if err != nil {
		t.Fatal(err)
	}
	customer1, err = GetCustomerById(customer1.GetID(), nil)
	log.Println("Version", customer1.GetVersion(), "FirstName", customer1.Person.FirstName)

	// Due to Rollback, FirstName should be "Foo" again
	if customer1.Person.FirstName != "Foo" {
		t.Fail()
	}
}

func create2CustomersAndPerformSomeUpserts(t *testing.T) (*Customer, *Customer) {
	DropAllCustomers()
	customer, err := NewCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Perform 3 Upserts
	customer.Person.FirstName = "Foo"
	err = customer.Upsert()
	customer.Person.MiddleName = "Bob"
	err = customer.Upsert()
	customer.Person.LastName = "Bar"
	err = customer.Upsert()
	address := &Address{
		Street:       "Holzweg",
		StreetNumber: "5",
	}
	err = customer.Upsert()
	// Create a second customer to make the history a little more interesting
	customer2, err := NewCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}
	customer2.Person.FirstName = "Trent"
	customer2.Upsert()
	customer2.Person.MiddleName = "The"
	customer2.Upsert()
	customer2.Person.LastName = "Second"
	customer2, err = customer2.UpsertAndGetCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}

	customer.AddAddress(address)
	err = customer.Upsert()
	if err != nil {
		t.Fatal(err)
	}
	customer.Person.FirstName = "Alice"
	customer.RemoveAddress(customer.GetAddresses()[0].GetID())
	customer, err = customer.UpsertAndGetCustomer(nil)
	if err != nil {
		t.Fatal(err)
	}
	return customer, customer2
}
