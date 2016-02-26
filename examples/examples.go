//Package examples hosts some common examples
package examples

import (
	"log"
	"time"

	"github.com/foomo/shop/order"
	"gopkg.in/mgo.v2/bson"
)

type SmurfOrderCustom struct {
	ResponsibleSmurf string
	Complete         bool
}

type SmurfPositionCustom struct {
	Foo string
}

type SmurfAddressCustom struct {
	Bar string
}

type SmurfCustomerCustom struct {
	FooBar string
}

// OrderCustom custom object provider
type SmurfOrderCustomProvider struct{}

type SmurfProcessor struct {
	Smurf          string
	CountProcessed int
	chanCount      chan int
}

func NewSmurfProcessor(name string) *SmurfProcessor {
	sp := &SmurfProcessor{
		Smurf:     name,
		chanCount: make(chan int),
	}
	go func() {
		for {
			select {
			case addCount := <-sp.chanCount:
				sp.CountProcessed += addCount
			}
		}
	}()
	return sp
}

func (sp *SmurfProcessor) Query() *bson.M {
	return &bson.M{"custom.responsiblesmurf": sp.Smurf}
}

func (sp *SmurfProcessor) Process(o *order.Order) error {
	sp.chanCount <- 1
	time.Sleep(time.Millisecond * 20)
	if sp.CountProcessed%100 == 0 {
		log.Println(sp.Smurf, sp.CountProcessed)
	}
	return nil
}

func (sp *SmurfProcessor) Concurrency() int {
	return 12
}

func (sp *SmurfProcessor) OrderCustomProvider() order.OrderCustomProvider {
	return &SmurfOrderCustomProvider{}
}

func (cp SmurfOrderCustomProvider) NewOrderCustom() interface{} {
	return &SmurfOrderCustom{}
}

func (cp SmurfOrderCustomProvider) NewPositionCustom() interface{} {
	return &SmurfPositionCustom{}
}

func (cp SmurfOrderCustomProvider) NewAddressCustom() interface{} {
	return &SmurfAddressCustom{}
}

func (cp SmurfOrderCustomProvider) NewCustomerCustom() interface{} {
	return &SmurfCustomerCustom{}
}

func (cp SmurfOrderCustomProvider) Fields() *bson.M {
	return nil
}
