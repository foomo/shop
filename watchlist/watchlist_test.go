package watchlist

import (
	"testing"

	"github.com/foomo/shop/test_utils"
	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
)

func TestWatchListsManipulate(t *testing.T) {
	test_utils.DropAllCollections()
	customerID := unique.GetNewID()
	_, err := NewCustomerWatchListsFromCustomerID(customerID)
	if err != nil {
		t.Fatal(err)
	}
	cw, err := GetCustomerWatchListsByCustomerID(customerID)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	// Create List
	listA, err := cw.AddList("TypeX", "ListA", true, "me", utils.GetDateYYYY_MM_DD(), "My awesome list")
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	// Add item to list
	err = cw.ListAddItem(listA.Id, "item1", 2)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	// Increase Quantity of item
	err = cw.ListAddItem(listA.Id, "item1", 3)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	// Add another item
	err = cw.ListAddItem(listA.Id, "item2", 2)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	item, err := cw.GetItem(listA.Id, "item1")
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	if item.Quantity != 5 {
		utils.PrintJSON(cw)
		t.Fatal("Wrong Quantity, expected 5")
	}
	// Reduce Quantity of item by 1
	err = cw.ListRemoveItem(listA.Id, "item2", 1)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	item, err = cw.GetItem(listA.Id, "item2")
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	if item.Quantity != 1 {
		t.Fatal("Wrong Quantity, expected 1")
	}
	// Remove last of item2
	err = cw.ListRemoveItem(listA.Id, "item2", 1)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}

	if len(listA.Items) != 1 {
		utils.PrintJSON(cw)
		t.Fatal("Expected 1 item in ListA")
	}
	newDescription := "new description"
	newName := "newName"
	// Edit list
	_, err = cw.EditList(listA.Id, newName, false, "", "", newDescription)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	if listA.Name != newName || listA.Description != newDescription || listA.PublicURIHash != "" {
		utils.PrintJSON(cw)
		t.Fatal("EditList failed")
	}
	// Set item fulfilled
	cw.ListSetItemFulfilled(listA.Id, "item1", 2)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	item, err = cw.GetItem(listA.Id, "item1")
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	if item.QtyFulfilled != 2 {
		utils.PrintJSON(cw)
		t.Fatal("Expected QtyFulfilled == 2")
	}

	// Create second CustomerWatchLists and merge
	sessionID := unique.GetNewID()
	_, err = NewCustomerWatchListsFromSessionID(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	cw2, err := GetCustomerWatchListsBySessionID(sessionID)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	listB, err := cw2.AddList("TypeX", "ListB", false, "recipient string", "targetDate string", "watchlist from session")
	if err != nil {
		t.Fatal(err)
	}
	err = cw2.ListAddItem(listB.Id, "item1b", 2)
	if err != nil {
		utils.PrintJSON(cw2)
		t.Fatal(err)
	}
	err = cw2.ListAddItem(listB.Id, "item1", 2)
	if err != nil {
		utils.PrintJSON(cw2)
		t.Fatal(err)
	}

	// Merge ListB from cw2 into ListA from cw
	err = MergeLists(cw2, listB.Id, cw, listA.Id)
	if err != nil {
		t.Fatal(err)
	}
	item, err = cw.GetItem(listA.Id, "item1")
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	if item.Quantity != 7 {
		utils.PrintJSON(cw)
		t.Fatal("Expected Quantity == 7")
	}
	item, err = cw2.GetItem(listB.Id, "item1b")
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}
	if item.Quantity != 2 {
		utils.PrintJSON(cw)
		t.Fatal("Expected Quantity == 2")
	}
	utils.PrintJSON(cw)

	// Test Getter
	cw, err = GetCustomerWatchListsByCustomerID(customerID)
	if err != nil {
		utils.PrintJSON(cw)
		t.Fatal(err)
	}

	// Test for non existant Id
	_, err = GetCustomerWatchListsByCustomerID("InvalidID")
	if err == nil {
		t.Fatal("Expected error not found")
	}

}
