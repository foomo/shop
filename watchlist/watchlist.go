package watchlist

import (
	"errors"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/unique"
	"github.com/foomo/shop/utils"
)

type CustomerWatchLists struct {
	BsonId     bson.ObjectId `bson:"_id,omitempty"`
	CustomerID string        `bson:"customerID"`
	SessionID  string        `bson:"sessionID"`
	Email      string        `bson:"email"`
	Lists      []*WatchList  `bson:"lists"`
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

func (cw *CustomerWatchLists) AddList(watchlistType string, name string, public bool, recipient string, targetDate string, description string) (*WatchList, error) {
	watchList := &WatchList{
		Id:            unique.GetNewID(),
		Type:          watchlistType,
		Name:          name,
		CreatedAt:     utils.GetDateYYYY_MM_DD(),
		Recipient:     recipient,
		TargetDate:    targetDate,
		Description:   description,
		PublicURIHash: "",
		Items:         []*WatchListItem{},
	}
	if public {
		watchList.PublicURIHash = unique.GetNewID()
	} else {
		watchList.PublicURIHash = ""
	}
	cw.Lists = append(cw.Lists, watchList)
	err := cw.Upsert()
	if err != nil {
		return nil, err
	}
	return watchList, nil
}
func (cw *CustomerWatchLists) RemoveListById(id string) error {
	listsTmp := []*WatchList{}
	for _, list := range cw.Lists {
		if list.Id != id {
			listsTmp = append(listsTmp, list)
		}
	}
	cw.Lists = listsTmp
	return cw.Upsert()
}

func (cw *CustomerWatchLists) ListAddItem(listId string, id string, quantity float64) error {
	if quantity < 1 {
		return errors.New("Quantity must not be smaller than 1")
	}
	listExists := false
	itemExists := false
	for _, list := range cw.Lists {
		if list.Id == listId {
			listExists = true
			for _, item := range list.Items {
				if item.Id == id {
					itemExists = true
					// Item already exists in list => increase Quantity
					item.Quantity = item.Quantity + quantity
				}
			}
			if !itemExists {
				list.Items = append(list.Items, &WatchListItem{
					Id:        id,
					DateAdded: utils.GetDateYYYY_MM_DD(),
					Quantity:  quantity,
				})
			}
		}
	}
	if !listExists {
		return errors.New("List with Id " + listId + " not found")
	}
	return cw.Upsert()
}

// ListRemoveItem will reduce the quantity of the item by one. if quantity equals or is less than zero, item is removed from list.
func (cw *CustomerWatchLists) ListRemoveItem(listId string, Id string, quantity float64) error {
	listExists := false
	for _, list := range cw.Lists {
		if list.Id == listId {
			listExists = true
			itemsTmp := []*WatchListItem{}
			for _, item := range list.Items {
				if item.Id == Id {
					item.Quantity = item.Quantity - quantity
					if item.Quantity > 0 {
						itemsTmp = append(itemsTmp, item)
					}
				} else {
					itemsTmp = append(itemsTmp, item)
				}
				list.Items = itemsTmp
			}
		}
	}
	if !listExists {
		return errors.New("List with Id " + listId + " not found")
	}
	return cw.Upsert()
}

func (cw *CustomerWatchLists) ListSetItemFulfilled(listId string, itemId string, quantity float64) error {
	listExists := false
	itemExists := false
	for _, list := range cw.Lists {
		if list.Id == listId {
			listExists = true
			for _, item := range list.Items {
				if item.Id == itemId {
					itemExists = true
					item.QtyFulfilled = item.QtyFulfilled + quantity
					return cw.Upsert()
				}
			}
		}
	}
	if !itemExists {
		return errors.New("Item with id " + itemId + " not found")
	}
	if !listExists {
		return errors.New("List with Id " + listId + " not found")
	}
	return cw.Upsert()
}
func (cw *CustomerWatchLists) GetList(listId string) (*WatchList, error) {
	for _, list := range cw.Lists {
		if list.Id == listId {
			return list, nil
		}
	}
	return nil, errors.New("List with Id " + listId + " not found")
}
func (cw *CustomerWatchLists) GetItem(listId string, itemId string) (*WatchListItem, error) {
	listExists := false
	for _, list := range cw.Lists {
		if list.Id == listId {
			listExists = true
			for _, item := range list.Items {
				if item.Id == itemId {
					return item, nil
				}
			}
		}
	}

	if !listExists {
		return nil, errors.New("List with Id " + listId + " not found")
	}
	return nil, errors.New("No item found for id " + itemId)
}
func (cw *CustomerWatchLists) GetListByURIHash(uriHash string) (*WatchList, error) {
	for _, list := range cw.Lists {
		if list.PublicURIHash == uriHash {
			return list, nil
		}
	}
	return nil, errors.New("No list found for URI hash")
}

func (cw *CustomerWatchLists) EditList(listId string, name string, public bool, recipient string, targetDate string, description string) (*WatchList, error) {
	for _, list := range cw.Lists {
		if list.Id == listId {
			if list.Name != "" {
				list.Name = name
			}
			if list.Recipient != "" {
				list.Recipient = recipient
			}
			if list.TargetDate != "" {
				list.TargetDate = targetDate
			}
			if list.Description != "" {
				list.Description = description
			}
			if public {
				if list.PublicURIHash == "" {
					list.PublicURIHash = unique.GetNewID()
				}
			} else {
				list.PublicURIHash = ""
			}
			err := cw.Upsert()
			if err != nil {
				return nil, err
			}
			return list, nil
		}
	}
	return nil, errors.New("List with Id " + listId + " not found")
}

// MergeLists merges a list from one CustomerWatchLists cwFrom Into a list of CustomerWatchLists cwInto
func MergeLists(cwFrom *CustomerWatchLists, listIdFrom string, cwInto *CustomerWatchLists, listIdInto string) error {
	// Find lists
	var listFrom *WatchList
	for _, list := range cwFrom.Lists {
		if list.Id == listIdFrom {
			listFrom = list
		}
	}
	if listFrom == nil {
		return errors.New("Could not find list with id" + listIdFrom)
	}

	// Merge
	for _, item := range listFrom.Items {
		err := cwInto.ListAddItem(listIdInto, item.Id, item.Quantity)
		if err != nil {
			return err
		}
	}
	return cwInto.Upsert()
}
