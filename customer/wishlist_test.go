package customer

import (
	"log"
	"testing"

	"github.com/foomo/shop/test_utils"
	"github.com/foomo/shop/utils"
)

func TestWishList(t *testing.T) {
	test_utils.DropAllCollections()
	log.Println("Wishlist...")
	customer, err := NewCustomer(MOCK_EMAIL, MOCK_PASSWORD, nil)
	if err != nil {
		t.Fatal(err)
	}
	idA, err := customer.WishListsAddList("List A", true, "Grandma", "2016-12-24", "Awesome description")
	if err != nil {
		t.Fatal(err)
	}
	idB, err := customer.WishListsAddList("List B", false, "GrandPa", "2016-12-24", "Awesome description")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListAddItem(idA, "ItemA")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListAddItem(idA, "ItemA2")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListAddItem(idA, "ItemA3")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListAddItem(idB, "ItemB")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListRemoveItem(idA, "ItemA2")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListsRemoveListById(idB)
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListSetAccess(idA, true)
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListSetItemFulfilled(idA, "ItemA", true)
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListSetRecipient(idA, "Brother")
	if err != nil {
		t.Fatal(err)
	}
	err = customer.WishListSetRecipient("idC", "Brother")
	if err == nil {
		t.Fatal(err)
	}
	err = customer.WishListSetItemFulfilled(idA, "ItemNotExisting", true)
	if err == nil {
		t.Fatal(err)
	}
	utils.PrintJSON(customer.GetWishLists())

}
