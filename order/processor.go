package order

import (
	"gopkg.in/mgo.v2/bson"
)

type Processor interface {
	OrderCustomProvider() OrderCustomProvider
	GetQuery() *bson.M
	SetQuery(*bson.M)
	Process(*Order) error
	Concurrency() int
}

type BulkProcessor interface {
	OrderCustomProvider() OrderCustomProvider
	ProcessBulk([]*Order) []error
	GetQuery() *bson.M
	SetQuery(*bson.M)
	Limit() int
	Concurrency() int
}
