package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Foo struct {
	Date  time.Time
	Bar   string
	Blubb string
}

type CustomProvider struct {
	/* implements OrderCustomProvider interface */
}

func (custom CustomProvider) NewOrderCustom() interface{} {
	return &Foo{}
}

func (custom CustomProvider) NewPositionCustom() interface{} {
	return &Foo{}
}

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

	cachedOrder, err = GetOrderCacheEntry(order.Id, nil)
	assert.Error(t, err, "cache hit but cache should be empty")
	assert.Nil(t, cachedOrder)

	err = SetOrderCacheEntry(order)
	assert.NoError(t, err, "unable to set cache entry")

	cachedOrder, err = GetOrderCacheEntry(order.Id, nil)
	assert.NoError(t, err, "cache hit failed")
	assert.NotNil(t, cachedOrder, "order still empty")
	assert.Equal(t, order.Id, cachedOrder.Id, "orderID mismatch")
	assert.Equal(t, order.ShopID, cachedOrder.ShopID, "shopID mismatch")

	foo := &Foo{
		Date:  time.Now(),
		Bar:   "Test",
		Blubb: "Lorem",
	}

	order.Custom = foo
	errUpsert := order.Upsert()
	assert.NoError(t, errUpsert)

	orderNew, errGetOrder := GetOrderById(order.Id, &CustomProvider{})

	assert.NoError(t, errGetOrder)
	assert.NotNil(t, orderNew)
	assert.Equal(t, order.Id, orderNew.Id)
	assert.Equal(t, foo.Bar, orderNew.Custom.(*Foo).Bar, "custom data mismatch")

	orderNew.Custom.(*Foo).Bar = "Blubb"

	version := orderNew.Version.GetVersion()

	errUpsert = orderNew.Upsert()
	assert.NoError(t, errUpsert)
	assert.Equal(t, "Blubb", orderNew.Custom.(*Foo).Bar, "custom data mismatch the second")

	assert.True(t, version+1 == orderNew.Version.Current, "version mismatch")
}
