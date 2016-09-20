package customer

import (
	"log"
	"testing"

	"github.com/foomo/shop/test_utils"
	"github.com/foomo/shop/utils"
)

func TestWatchList(t *testing.T) {
	test_utils.DropAllCollections()
	log.Println("Wishlist...")
	customer, err := NewCustomer(MOCK_EMAIL, MOCK_PASSWORD, nil)
	if err != nil {
		t.Fatal(err)
	}
	lists, err := customer.WatchListsAddList("TypeA", "List A", true, "Grandma", "2016-12-24", "Awesome description")
	if err != nil {
		t.Fatal(err)
	}
	idA := lists[len(lists)-1].Id
	lists, err = customer.WatchListsAddList("TypeB", "List B", false, "GrandPa", "2016-12-24", "Awesome description")
	if err != nil {
		t.Fatal(err)
	}
	idB := lists[len(lists)-1].Id
	err = customer.WatchListAddItem(idA, "ItemA")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListAddItem(idA, "ItemA2")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListAddItem(idA, "ItemA3")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListAddItem(idB, "ItemB")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListRemoveItem(idA, "ItemA2")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListsRemoveListById(idB)
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListSetAccess(idA, true)
	if err != nil {
		t.Fatal(err)
	}

	err = customer.WatchListSetRecipient(idA, "Brother")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WatchListSetRecipient("idC", "Brother")
	if err == nil {
		t.Fatal(err)
	}

	utils.PrintJSON(customer.GetWatchLists())

}
