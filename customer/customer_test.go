package customer

import (
	"log"
	"testing"
)

const (
	MOCK_EMAIL            = "Foo@Bar.com"
	MOCK_PASSWORD         = "supersafepassword!11"
	MOCK_EMAIL2           = "Alice@Bar.com"
	MOCK_PASSWORD2        = "evensaferpassword!11!§$%&"
	OPEN_DIFFS_IN_BROWSER = false
)

func TestCustomerGetLatestCustomerFromDb(t *testing.T) {
	DropAllCustomersAndCredentials()
	customer, err := NewCustomer(MOCK_EMAIL, MOCK_PASSWORD, nil)
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
	customer, err = GetCurrentCustomerByIdFromVersionsHistory(customer.GetID(), nil)
	if customer.GetVersion().Current != 3 {
		log.Println("Version is ", customer.GetVersion().Current, "- should have been 3.")
		t.Fail()
	}

}
func TestCustomerDiff2LatestCustomerVersions(t *testing.T) {
	customer1, _ := create2CustomersAndPerformSomeUpserts(t)

	_, err := DiffTwoLatestCustomerVersions(customer1.GetID(), nil, OPEN_DIFFS_IN_BROWSER)
	if err != nil {
		t.Fatal(err)
	}
}
func TestCustomerRollbackAndDiff(t *testing.T) {
	customer1, _ := create2CustomersAndPerformSomeUpserts(t)

	errRoll := customer1.Rollback(customer1.GetVersion().Current - 1)
	if errRoll != nil {
		t.Fatal(errRoll)
	}
	customer1, errRoll = GetCustomerById(customer1.GetID(), nil)

	_, err := DiffTwoLatestCustomerVersions(customer1.GetID(), nil, OPEN_DIFFS_IN_BROWSER)
	if err != nil {
		t.Fatal(err)
	}
}
func TestCustomerRollback(t *testing.T) {
	customer1, _ := create2CustomersAndPerformSomeUpserts(t)
	log.Println("Version", customer1.GetVersion(), "FirstName", customer1.Person.FirstName)
	err := customer1.Rollback(customer1.GetVersion().Current - 1)
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
	DropAllCustomersAndCredentials()
	customer, err := NewCustomer(MOCK_EMAIL, MOCK_PASSWORD, nil)
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
	customer2, err := NewCustomer(MOCK_EMAIL2, MOCK_PASSWORD2, nil)
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
