package order

import (
	"gopkg.in/mgo.v2/bson"
)

type Processor interface {
	OrderCustomProvider() OrderCustomProvider
	Query() *bson.M
	Process() error
	Concurrency() int
}
