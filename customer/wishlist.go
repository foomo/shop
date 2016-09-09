package customer

import (
	"errors"

	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
)

type WishLists struct {
	Lists []*WishList
}

type WishList struct {
	Id          string
	ListName    string
	CreatedAt   string
	TargetDate  string
	Description string
	Recipient   string
	Public      bool
	Items       []*WishListItem
}

type WishListItem struct {
	ItemId    string
	DateAdded string
	Fulfilled bool
}

func (customer *Customer) GetWishLists() *WishLists {
	return customer.WishLists
}
func (customer *Customer) WishListsAddList(name string, public bool, recipient string, targetDate string, description string) (id string, err error) {
	wishlist := &WishList{
		Id:          unique.GetNewID(),
		ListName:    name,
		CreatedAt:   utils.GetDateYYYY_MM_DD(),
		Recipient:   recipient,
		TargetDate:  targetDate,
		Description: description,
	}
	customer.GetWishLists().Lists = append(customer.GetWishLists().Lists, wishlist)
	return wishlist.Id, customer.Upsert()
}
func (customer *Customer) WishListsRemoveListById(id string) error {
	listsTmp := []*WishList{}
	for _, list := range customer.GetWishLists().Lists {
		if list.Id != id {
			listsTmp = append(listsTmp, list)
		}
	}
	customer.GetWishLists().Lists = listsTmp
	return customer.Upsert()
}

func (customer *Customer) WishListAddItem(listId string, itemId string) error {
	listExists := false
	for _, list := range customer.GetWishLists().Lists {
		if list.Id == listId {
			listExists = true
			list.Items = append(list.Items, &WishListItem{
				ItemId:    itemId,
				DateAdded: utils.GetDateYYYY_MM_DD(),
			})
		}
	}
	if !listExists {
		return errors.New("List with Id " + listId + " not found")
	}
	return customer.Upsert()
}
func (customer *Customer) WishListRemoveItem(listId string, itemId string) error {
	listExists := false
	for _, list := range customer.GetWishLists().Lists {
		if list.Id == listId {
			listExists = true
			itemsTmp := []*WishListItem{}
			for _, item := range list.Items {
				if item.ItemId != itemId {
					itemsTmp = append(itemsTmp, item)
				}
			}
			list.Items = itemsTmp
		}
	}
	if !listExists {
		return errors.New("List with Id " + listId + " not found")
	}
	return customer.Upsert()
}
func (customer *Customer) WishListSetItemFulfilled(listId string, itemId string, fulfilled bool) error {
	item, err := customer.GetWishLists().findItem(listId, itemId)
	if err != nil {
		return err
	}
	item.Fulfilled = fulfilled

	return customer.Upsert()
}

func (customer *Customer) WishListSetRecipient(listId string, recipient string) error {
	for _, list := range customer.GetWishLists().Lists {
		if list.Id == listId {
			list.Recipient = recipient
			return customer.Upsert()
		}
	}
	return errors.New("List with Id " + listId + " not found")
}
func (customer *Customer) WishListSetAccess(listId string, public bool) error {

	for _, list := range customer.GetWishLists().Lists {
		if list.Id == listId {
			list.Public = public
			return customer.Upsert()
		}
	}
	return errors.New("List with Id " + listId + " not found")
}

func (w *WishLists) findItem(listId string, itemId string) (*WishListItem, error) {

	listExists := false
	for _, list := range w.Lists {
		if list.Id == listId {
			listExists = true
			for _, item := range list.Items {
				if item.ItemId == itemId {
					return item, nil
				}
			}
		}
	}
	if !listExists {
		return nil, errors.New("List with Id " + listId + " not found")
	}

	return nil, errors.New("Item with Id " + itemId + " not found")

}
