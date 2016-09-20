package customer

import (
	"errors"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
)

type WatchLists struct {
	BsonId     bson.ObjectId `bson:"_id,omitempty"`
	CustomerID string
	Lists      []*WatchList
}

type WatchList struct {
	Id            string
	Type          string
	Name          string
	CreatedAt     string
	TargetDate    string
	Description   string
	Recipient     string
	PublicURIHash string
	Items         []*WatchListItem
}

type WatchListItem struct {
	Id           string
	DateAdded    string
	Quantity     float64
	QtyFulfilled float64
}

func (customer *Customer) GetWatchLists() []*WatchList {
	return customer.WatchLists.Lists
}
func (customer *Customer) WatchListsAddList(watchlistType string, name string, public bool, recipient string, targetDate string, description string) ([]*WatchList, error) {
	watchList := &WatchList{
		Id:          unique.GetNewID(),
		Type:        watchlistType,
		Name:        name,
		CreatedAt:   utils.GetDateYYYY_MM_DD(),
		Recipient:   recipient,
		TargetDate:  targetDate,
		Description: description,
	}
	customer.WatchLists.Lists = append(customer.WatchLists.Lists, watchList)
	err := customer.Upsert()
	if err != nil {
		return nil, err
	}
	return customer.GetWatchLists(), nil
}
func (customer *Customer) WatchListsRemoveListById(id string) error {
	listsTmp := []*WatchList{}
	for _, list := range customer.GetWatchLists() {
		if list.Id != id {
			listsTmp = append(listsTmp, list)
		}
	}
	customer.WatchLists.Lists = listsTmp
	return customer.Upsert()
}

func (customer *Customer) WatchListAddItem(listId string, Id string) error {
	listExists := false
	for _, list := range customer.GetWatchLists() {
		if list.Id == listId {
			listExists = true
			list.Items = append(list.Items, &WatchListItem{
				Id:        Id,
				DateAdded: utils.GetDateYYYY_MM_DD(),
			})
		}
	}
	if !listExists {
		return errors.New("List with Id " + listId + " not found")
	}
	return customer.Upsert()
}
func (customer *Customer) WatchListRemoveItem(listId string, Id string) error {
	listExists := false
	for _, list := range customer.GetWatchLists() {
		if list.Id == listId {
			listExists = true
			itemsTmp := []*WatchListItem{}
			for _, item := range list.Items {
				if item.Id != Id {
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

// func (customer *Customer) WatchListSetItemFulfilled(listId string, itemId string, fulfilled bool) error {
// 	item, err := customer.WatchLists.findItem(listId, itemId)
// 	if err != nil {
// 		return err
// 	}
// 	item.Fulfilled = fulfilled

// 	return customer.Upsert()
// }

func (customer *Customer) WatchListSetRecipient(listId string, recipient string) error {
	for _, list := range customer.GetWatchLists() {
		if list.Id == listId {
			list.Recipient = recipient
			return customer.Upsert()
		}
	}
	return errors.New("List with Id " + listId + " not found")
}
func (customer *Customer) WatchListSetAccess(listId string, public bool) error {

	for _, list := range customer.GetWatchLists() {
		if list.Id == listId {
			if public {
				list.PublicURIHash = unique.GetNewID()
			} else {
				list.PublicURIHash = ""
			}
			return customer.Upsert()
		}
	}
	return errors.New("List with Id " + listId + " not found")
}

func (w *WatchLists) findItem(listId string, id string) (*WatchListItem, error) {

	listExists := false
	for _, list := range w.Lists {
		if list.Id == listId {
			listExists = true
			for _, item := range list.Items {
				if item.Id == id {
					return item, nil
				}
			}
		}
	}
	if !listExists {
		return nil, errors.New("List with Id " + listId + " not found")
	}

	return nil, errors.New("Item with Id " + id + " not found")

}
