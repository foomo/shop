package order

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/vmihailenco/msgpack"
)

var cache *bigcache.BigCache

func GetCache() (bg *bigcache.BigCache, err error) {

	if cache != nil {
		return cache, nil
	}

	config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 1024,
		// time after which entry can be evicted
		LifeWindow: 15 * time.Minute,
		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 1000 * 10 * 60,
		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 8192,
		// prints information about additional memory allocation
		Verbose: false,
		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 8192,
		// callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,
		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// Ignored if OnRemove is specified.
		OnRemoveWithReason: nil,
	}

	cacheInstance, errInitCache := bigcache.NewBigCache(config)
	if errInitCache != nil {
		err = errInitCache
		return
	}

	cache = cacheInstance
	bg = cacheInstance
	return
}

func GetOrderCacheEntry(orderID string) (order *Order, err error) {
	c, errCache := GetCache()
	if errCache != nil {
		err = errCache
		return
	}

	orderBytes, errCacheHit := c.Get(orderID)
	if errCacheHit != nil {
		err = errCacheHit
		return
	}

	errUnmarshal := msgpack.Unmarshal(orderBytes, &order)
	if errUnmarshal != nil {
		err = errUnmarshal
		return
	}

	return
}

func SetOrderCacheEntry(order *Order) error {
	c, errCache := GetCache()
	if errCache != nil {
		return errCache
	}

	orderBytes, errMarshall := msgpack.Marshal(order)
	if errMarshall != nil {
		return errMarshall
	}

	return c.Set(order.Id, orderBytes)
}

func RemoveOrderCacheEntry(orderID string) error {
	c, errCache := GetCache()
	if errCache != nil {
		return errCache
	}

	return c.Delete(orderID)
}
