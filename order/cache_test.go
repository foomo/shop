package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderCache(t *testing.T) {

	order, errOrder := NewOrderWithCustomId(nil, func() (string, error) {
		return "0815", nil
	})
	assert.NoError(t, errOrder)

	order.ShopID = "test"

	var cachedOrder *Order
	var err error

	// clear cache, since NewOrderWithCustomId gets the order at the end which automatically warms the cache
	errRemove := RemoveOrderCacheEntry(order.Id)
	assert.NoError(t, errRemove, "cache cleanup failed")

	assert.NotEmpty(t, order.Id, "unexpected empty orderID")
	assert.Equal(t, "0815", order.Id, "unexpected orderID")

	cachedOrder, err = GetOrderCacheEntry(order.Id)
	assert.Error(t, err, "cache hit but cache should be empty")
	assert.Nil(t, cachedOrder)

	err = SetOrderCacheEntry(order)
	assert.NoError(t, err, "unable to set cache entry")

	cachedOrder, err = GetOrderCacheEntry(order.Id)
	assert.NoError(t, err, "cache hit failed")
	assert.NotNil(t, cachedOrder, "order still empty")
	assert.Equal(t, order.Id, cachedOrder.Id, "orderID mismatch")
	assert.Equal(t, order.ShopID, cachedOrder.ShopID, "shopID mismatch")
}
