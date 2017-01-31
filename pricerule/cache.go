package pricerule

import (
	"errors"
	"sync"

	"gopkg.in/mgo.v2/bson"
)

// Cache - the cache struct type
type Cache struct {
	groupsCache          map[GroupType]map[string][]string
	loadGroupsCacheMutex *sync.Mutex
	loadPriceRulesMutex  *sync.Mutex
	validPriceRules      []PriceRule
	enabled              bool
}

// NewCache -
func NewCache() *Cache {
	c := &Cache{}
	c.loadGroupsCacheMutex = &sync.Mutex{}

	c.enabled = false
	return c
}

// GetGroupsCache -
func (c *Cache) GetGroupsCache() map[GroupType]map[string][]string {
	return c.groupsCache

}

// InitCache - load groups data into memory
func (c *Cache) InitCache() error {
	return cache.loadGroupCacheByItem()
}

// ClearCache - will force the use of data from db
func (c *Cache) ClearCache() {
	c.groupsCache = make(map[GroupType]map[string][]string)
	c.enabled = false
}

// CacheAddGroupToItems -
func (c *Cache) CacheAddGroupToItems(itemIDs []string, groupID string, groupType GroupType) error {
	//synchronize code code
	if !c.enabled {
		return nil
	}
	c.loadGroupsCacheMutex.Lock()
	defer cache.loadGroupsCacheMutex.Unlock()

	for _, itemID := range itemIDs {
		if _, ok := c.groupsCache[groupType]; ok {

			if _, ok := cache.groupsCache[groupType][itemID]; ok {
				c.groupsCache[groupType][itemID] = RemoveDuplicates(append(cache.groupsCache[groupType][itemID], groupID))
			} else {
				c.groupsCache[groupType][itemID] = []string{groupID}
			}
		} else {
			// only if cached modify cached
			c.groupsCache[groupType] = make(map[string][]string)
			c.groupsCache[groupType][itemID] = []string{groupID}
		}

	}
	return nil
}

// CacheDeleteGroup -
func (c *Cache) CacheDeleteGroup(group *Group) error {
	if !c.enabled {
		return nil
	}
	//synchronize code code
	c.loadGroupsCacheMutex.Lock()
	defer cache.loadGroupsCacheMutex.Unlock()

	if group != nil {
		if _, ok := c.groupsCache[group.Type]; ok {
			for _, itemID := range group.ItemIDs {
				if _, ok := c.groupsCache[group.Type][itemID]; ok {
					c.groupsCache[group.Type][itemID] = removeValueFromArray(group.ID, c.groupsCache[group.Type][itemID])
				}
			}
		}
		// the group type is not in the cache anyway
		return nil
	}
	return errors.New("nil group passed to CacheDeleteGroup")
}

// remove value from array - by value not index
func removeValueFromArray(val string, vals []string) []string {
	ret := []string{}
	for _, inVal := range vals {
		if inVal != val {
			ret = append(ret, inVal)
		}
	}
	return ret
}

func (c *Cache) loadGroupCacheByItem() error {
	//synchronize code code
	c.loadGroupsCacheMutex.Lock()
	defer c.loadGroupsCacheMutex.Unlock()
	tempMap := make(map[GroupType]map[string][]string)
	for _, groupType := range []GroupType{ProductGroup, CustomerGroup} {
		p := GetPersistorForObject(new(Group))
		query := bson.M{"type": groupType}
		var result []struct {
			ID      string   `bson:"id"`
			Name    string   `bson:"name"`
			itemIDs []string `bson:"itemids"`
		}

		err := p.GetCollection().Find(query).Select(bson.M{"id": 1}).Sort("priority").All(&result)
		if err != nil {
			return err
		}

		if _, ok := tempMap[groupType]; ok {
			tempMap[groupType] = make(map[string][]string)
		}

		for _, group := range result {
			for _, itemID := range group.itemIDs {

				// if we have a key already append
				if _, ok := tempMap[groupType][itemID]; ok {
					tempMap[groupType][itemID] = append(tempMap[groupType][itemID], group.ID)
				} else {
					// initialize and add
					tempMap[groupType][itemID] = []string{group.ID}
				}
			}
		}

	}
	c.groupsCache = tempMap
	c.enabled = true
	return nil
}
