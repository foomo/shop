package pricerule

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Cache - the cache struct type
type Cache struct {
	groupsCache            map[GroupType]map[string][]string
	catalogValidRulesCache []PriceRule
	cacheMutex             *sync.Mutex
	blacklistedItemIDs     []string

	enabled bool
}

// NewCache -
func NewCache() *Cache {
	c := &Cache{}
	c.cacheMutex = &sync.Mutex{}
	c.enabled = false
	return c
}

// GetGroupsCache -
func (c *Cache) GetGroupsCache() map[GroupType]map[string][]string {
	return c.groupsCache
}

// InitCatalogCalculationCache - load groups data into memory
func (c *Cache) InitCatalogCalculationCache() (err error) {
	now := time.Now()
	//synchronize code code
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	err = c.loadGroupCacheByItem()
	if err != nil {
		return
	}

	catalogValidRulesCache, getPromotionsErr := GetValidPriceRulesForPromotions([]Type{TypePromotionCustomer, TypePromotionProduct, TypePromotionOrder}, nil)
	if getPromotionsErr != nil {
		err = getPromotionsErr
		return
	}
	c.catalogValidRulesCache = catalogValidRulesCache

	//load the blacklisted products
	blacklistedItems, blacklistLoadErr := GetBlacklistedItemIds()
	if blacklistLoadErr != nil {
		err = blacklistLoadErr
		log.Println(err)
		return
	}
	c.blacklistedItemIDs = blacklistedItems
	// done all
	c.enabled = true
	timeTrack(now, "[InitCatalogCalculationCache] cache loading took ")

	return
}

// ClearCache - will force the use of data from db
func (c *Cache) ClearCatalogCalculationCache() {
	c.groupsCache = make(map[GroupType]map[string][]string)
	c.catalogValidRulesCache = []PriceRule{}
	c.enabled = false
}

// GetGroupsIDSForItem - use cache or fallback to db retrieval
func (c *Cache) GetGroupsIDSForItem(itemID string, groupType GroupType) []string {

	// if we have the cache use it,
	if c.enabled == true {
		if groupIDs, ok := c.groupsCache[groupType][itemID]; ok {
			return groupIDs
		} else {
			groupIDs := []string{} //TODO: if not in the cache means we do not have it
			c.groupsCache[groupType][itemID] = groupIDs
			return groupIDs
		}
	}
	groupIDs := GetGroupsIDSForItem(itemID, groupType)
	return groupIDs
}

// CachedGetValidProductAndCustomerPriceRules -
func (c *Cache) CachedGetValidProductAndCustomerPriceRules(customProvider PriceRuleCustomProvider) ([]PriceRule, error) {
	if c.enabled {
		return c.catalogValidRulesCache, nil
	}
	return GetValidPriceRulesForPromotions([]Type{TypePromotionCustomer, TypePromotionProduct}, customProvider)
}

func (c *Cache) GetBlacklistedItemIDs() (itemIDs []string) {
	itemIDs = []string{}
	if c.enabled {
		return c.blacklistedItemIDs
	}
	itemIDs, err := GetBlacklistedItemIds()
	if err != nil {
		log.Println(err)
		itemIDs = []string{}
	}
	return
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
	now := time.Now()

	tempMap := make(map[GroupType]map[string][]string)
	for _, groupType := range []GroupType{ProductGroup, CustomerGroup} {
		p := GetPersistorForObject(&Group{})
		query := bson.M{"type": groupType}

		var result = []Group{}

		err := p.GetCollection().Find(query).Sort("priority").All(&result)
		if err != nil {
			return err
		}

		if _, ok := tempMap[groupType]; !ok {
			tempMap[groupType] = make(map[string][]string)
		}

		for _, group := range result {
			for _, itemID := range group.ItemIDs {
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
	if Verbose {
		fmt.Println(c.GetGroupsCache())
	}
	c.enabled = true
	timeTrack(now, "[loadGroupCacheByItem] cache loading took")

	return nil
}
