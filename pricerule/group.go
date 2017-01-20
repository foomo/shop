package pricerule

import (
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------
const (
	CustomerGroup GroupType = "customer-group"
	ProductGroup  GroupType = "product-group"
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
	ID             string        //group id - referenced by PriceRule (s)
	Name           string        //group name
	IDs            []string      //list of product IDs or customer IDs in assigned to the group
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Custom         interface{} `bson:",omitempty"` //make it extensible if needed
}

type emptyGroupType struct {
	Type           GroupType
	BsonID         bson.ObjectId `bson:"_id,omitempty"`
	ID             string        //group id - referenced by PriceRule (s)
	Name           string        //group name
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Custom         interface{} `bson:",omitempty"` //make it extensible if needed
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
	group.IDs = []string{}
	return true
}

// AddGroupIDsAndPersist - appends removes duplicates and persists
func (group *Group) AddGroupIDsAndPersist(itemIDs []string) bool {
	group.AddGroupItemIDs(itemIDs)

	//addtoset
	p := GetPersistorForObject(group) //GetGroupPersistor()
	_, err := p.GetCollection().Upsert(bson.M{"id": group.ID}, group)
	if err != nil {
		return false
	}

	return true
}

// AddGroupItemIDs - appends removes duplicates and persists
func (group *Group) AddGroupItemIDs(itemIDs []string) bool {
	var ids = append(group.IDs, itemIDs...)
	group.IDs = RemoveDuplicates(ids)
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
	err := GetPersistorForObject(new(Group)).GetCollection().EnsureIndex(index)
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
	p := GetPersistorForObject(group)
	if groupFromDb == nil {
		_, err = p.GetCollection().Upsert(bson.M{"id": group.ID}, group)
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

		_, err = p.GetCollection().Upsert(bson.M{"id": group.ID}, emptyCopy)
		if err != nil {
			return err
		}

		//make sure there are no duplicateas - $addToSet
		err = p.GetCollection().Update(bson.M{"id": group.ID}, bson.M{"$addToSet": bson.M{"itemids": bson.M{"$each": group.IDs}}})
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
	err := GetPersistorForObject(group).GetCollection().Remove(bson.M{"id": group.ID})
	group = nil
	return err
}

// DeleteGroup -
func DeleteGroup(ID string) error {
	err := GetPersistorForObject(new(Group)).GetCollection().Remove(bson.M{"id": ID})
	return err
}

// RemoveAllGroups -
func RemoveAllGroups() error {
	p := GetPersistorForObject(new(Group))
	_, err := p.GetCollection().RemoveAll(bson.M{})
	return err
}

// GetGroupsIDSForItem -
func GetGroupsIDSForItem(itemID string, groupType GroupType) []string {
	p := GetPersistorForObject(new(Group))
	query := bson.M{"itemids": bson.M{"$in": []string{itemID}}, "type": groupType}

	var ret = []string{}

	var result []struct {
		ID string `bson:"id"`
	}

	err := p.GetCollection().Find(query).Select(bson.M{"id": 1}).Sort("priority").All(&result)
	if err != nil {
		// handle error
		return []string{}
	}

	for _, val := range result {
		ret = append(ret, val.ID)
	}

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
