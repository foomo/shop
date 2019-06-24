package order

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"time"

	"github.com/allegro/bigcache"
	"github.com/foomo/shop/order"
	"github.com/foomo/shop/state"
	"github.com/foomo/shop/version"
)

var cache *bigcache.BigCache

func init() {

	log.Println("init order cache")

	gob.Register(time.Time{})
	gob.Register(order.Position{})
	gob.Register(order.CustomerData{})
	gob.Register(order.Processing{})
	gob.Register(state.State{})
	gob.Register(order.Flags{})
	gob.Register(version.Version{})
	gob.Register(order.Order{})

	config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 1024,
		// time after which entry can be evicted
		LifeWindow: 15 * time.Minute,
		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 1000 * 10 * 60,
		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 1024,
		// prints information about additional memory allocation
		Verbose: true,
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

	var errInitCache error
	cache, errInitCache = bigcache.NewBigCache(config)
	if errInitCache != nil {
		log.Fatal(errInitCache)
	}
}

func GetOrderCacheEntry(orderID string) (order *Order, err error) {
	entryBytes, errCacheHit := cache.Get(orderID)
	if errCacheHit != nil {
		err = errCacheHit
		return
	}

	var buffer bytes.Buffer

	_, errWrite := buffer.Write(entryBytes)
	if errWrite != nil {
		err = errWrite
		return
	}

	dec := gob.NewDecoder(&buffer)
	errDecode := dec.Decode(&order)
	if errDecode != nil {
		err = errDecode
		return
	}

	return
}

func SetOrderCacheEntry(order *Order) error {

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	errEncode := enc.Encode(order)
	if errEncode != nil {
		return errEncode
	}

	reader := bytes.NewReader(buffer.Bytes())
	entryBytes, errEntry := ioutil.ReadAll(reader)
	if errEntry != nil {
		return errEntry
	}

	return cache.Set(order.Id, entryBytes)
}

func RemoveOrderCacheEntry(orderID string) error {
	return cache.Delete(orderID)
}
