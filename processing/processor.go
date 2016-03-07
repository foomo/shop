package processing

import (
	"github.com/foomo/shop/order"
	"gopkg.in/mgo.v2/bson"
)

type Processor interface {
	OrderCustomProvider() order.OrderCustomProvider
	GetQuery() *bson.M
	SetQuery(*bson.M)
	Process(*order.Order) error
	Concurrency() int
}

type BulkProcessor interface {
	OrderCustomProvider() order.OrderCustomProvider
	ProcessBulk([]*order.Order) []error
	GetQuery() *bson.M
	SetQuery(*bson.M)
	Limit() int
	Concurrency() int
}
