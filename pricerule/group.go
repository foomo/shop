package pricerule

import (
	"time"

	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------
const (
	CustomerGroup  GroupType = "customer-group"
	ProductGroup   GroupType = "product-group"
	BlacklistGroup GroupType = "blacklist-group"
)

// GroupType - product group or customer group
type GroupType string

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

//Group - used to assign products to price rules
type Group struct {
	Type           GroupType
	BsonID         bson.ObjectId `bson:"_id,omitempty"`
	ID             string        `bson:"id"`      //group id - referenced by PriceRule (s)
	Name           string                         //group name
	ItemIDs        []string      `bson:"itemids"` //list of product IDs or customer IDs in assigned to the group
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Custom         interface{}   `bson:",omitempty"` //make it extensible if needed
}

type emptyGroupType struct {
	Type           GroupType
	BsonID         bson.ObjectId `bson:"_id,omitempty"`
	ID             string //group id - referenced by PriceRule (s)
	Name           string //group name
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Custom         interface{}   `bson:",omitempty"` //make it extensible if needed
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// LoadGroup -
func LoadGroup(ID string, customProvider PriceRuleCustomProvider) (*Group, error) {
	return GetGroupByID(ID, customProvider)
}

// RemoveAllProductIds - clear all product IDs
func (group *Group) RemoveAllProductIds() bool {
	group.ItemIDs = []string{}
	return true
}

// AddGroupItemIDsAndPersist - appends removes duplicates and persists
func (group *Group) AddGroupItemIDsAndPersist(itemIDs []string) bool {
	group.AddGroupItemIDs(itemIDs)

	//addtoset
	session, collection := GetPersistorForObject(group).GetCollection() //GetGroupPersistor()
	defer session.Close()

	_, err := collection.Upsert(bson.M{"id": group.ID}, group)
	if err != nil {
		return false
	}

	return true
}

// AddGroupItemIDs - appends removes duplicates and persists
func (group *Group) AddGroupItemIDs(itemIDs []string) bool {
	var ids = append(group.ItemIDs, itemIDs...)
	group.ItemIDs = RemoveDuplicates(ids)
	return true
}

// GroupAlreadyExistsInDB checks if a Group with given ID already exists in the database
func GroupAlreadyExistsInDB(ID string) (bool, error) {
	return ObjectOfTypeAlreadyExistsInDB(ID, new(Group))
}

// Upsert - upsers a group
// note that if you programmatically manipulate the CreatedAt time, this methd will upsert it
func (group *Group) Upsert() error {

	index := mgo.Index{
		Key:        []string{"itemids"},
		Unique:     false, // other froups can contain the same items !!!
		DropDups:   false, // other froups can contain the same items !!!
		Background: true,  // See notes.
		Sparse:     true,
	}
	session, collection := GetPersistorForObject(new(Group)).GetCollection() //GetGroupPersistor()
	defer session.Close()

	err := collection.EnsureIndex(index)
	var groupFromDb *Group
	//set created and modified times
	if group.CreatedAt.IsZero() {
		groupFromDb, err = GetGroupByID(group.ID, nil)
		if err != nil || groupFromDb == nil {
			group.CreatedAt = time.Now()
		} else {
			group.CreatedAt = groupFromDb.CreatedAt
		}
	}
	group.LastModifiedAt = time.Now()
	objectSession, objectCollection := GetPersistorForObject(group).GetCollection()
	defer objectSession.Close()

	if groupFromDb == nil {
		_, err = objectCollection.Upsert(bson.M{"id": group.ID}, group)
	} else {

		emptyCopy := emptyGroupType{
			ID:             group.ID,
			Name:           group.Name,
			BsonID:         group.BsonID,
			CreatedAt:      group.CreatedAt,
			LastModifiedAt: group.LastModifiedAt,
			Custom:         group.Custom,
			Type:           group.Type,
		}

		_, err = objectCollection.Upsert(bson.M{"id": group.ID}, emptyCopy)
		if err != nil {
			return err
		}

		//make sure there are no duplicateas - $addToSet
		err = objectCollection.Update(bson.M{"id": group.ID}, bson.M{"$addToSet": bson.M{"itemids": bson.M{"$each": group.ItemIDs}}})
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil

}

// Delete - delete group - ID must be set
func (group *Group) Delete() error {
	session, collection := GetPersistorForObject(group).GetCollection()
	defer session.Close()

	err := collection.Remove(bson.M{"id": group.ID})
	if err != nil {
		return err
	}
	group = nil
	return nil
}

// DeleteGroup -
func DeleteGroup(ID string) error {
	session, collection := GetPersistorForObject(new(Group)).GetCollection()
	defer session.Close()

	err := collection.Remove(bson.M{"id": ID})
	if err != nil {
		return err
	}
	return nil
}

// RemoveAllGroups -
func RemoveAllGroups() error {
	session, collection := GetPersistorForObject(new(Group)).GetCollection()
	defer session.Close()

	_, err := collection.RemoveAll(bson.M{})
	if err != nil {
		return err
	}
	return nil
}

// GetGroupsIDSForItem -
func GetGroupsIDSForItem(itemID string, groupType GroupType) []string {
	//now := time.Now()

	session, collection := GetPersistorForObject(new(Group)).GetCollection()
	defer session.Close()

	query := bson.M{"itemids": bson.M{"$in": []string{itemID}}, "type": groupType}

	var ret = []string{}

	var result []struct {
		ID string `bson:"id"`
	}

	err := collection.Find(query).Select(bson.M{"id": 1}).Sort("priority").All(&result)
	if err != nil {
		// handle error
		return []string{}
	}

	for _, val := range result {
		ret = append(ret, val.ID)
	}
	//timeTrack(now, "[GetGroupsIDSForItem -> from db]")
	return ret
}

// RemoveDuplicates - cleanup array
func RemoveDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

// GetBlacklistedItemIds -
func GetBlacklistedItemIds() (itemIDs []string, err error) {
	return getItemIDsFroGroupType(BlacklistGroup)
}

func getItemIDsFroGroupType(groupType GroupType) (itemIDs []string, err error) {
	session, collection := GetPersistorForObject(&Group{}).GetCollection()
	defer session.Close()

	query := bson.M{"type": groupType}
	var result = []Group{}
	findErr := collection.Find(query).Sort("priority").All(&result)
	if findErr != nil {
		log.Println(findErr)
		err = findErr
		return
	}
	for _, group := range result {
		if len(group.ItemIDs) > 0 {
			itemIDs = append(itemIDs, group.ItemIDs...)
		}
	}
	itemIDs = RemoveDuplicates(itemIDs)
	return
}
