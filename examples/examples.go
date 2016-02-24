//Package examples hosts some common examples
package examples

import (
	"log"

	"github.com/foomo/shop/order"
	"gopkg.in/mgo.v2/bson"
)

type OrderCustom struct {
	ResponsibleSmurf string
	Complete         bool
}

type PositionCustom struct {
	Foo string
}

type AddressCustom struct {
	Bar string
}

type CustomerCustom struct {
	FooBar string
}

// OrderCustom custom object provider
type FullOrderCustomProvider struct{}

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
	//time.Sleep(time.Millisecond * 20)
	if sp.CountProcessed%100 == 0 {
		log.Println(sp.Smurf, sp.CountProcessed)
	}
	return nil
}

func (sp *SmurfProcessor) Concurrency() int {
	return 12
}

func (sp *SmurfProcessor) OrderCustomProvider() order.OrderCustomProvider {
	return &FullOrderCustomProvider{}
}

func (cp FullOrderCustomProvider) NewOrderCustom() interface{} {
	return &OrderCustom{}
}

func (cp FullOrderCustomProvider) NewPositionCustom() interface{} {
	return &PositionCustom{}
}

func (cp FullOrderCustomProvider) NewAddressCustom() interface{} {
	return &AddressCustom{}
}

func (cp FullOrderCustomProvider) NewCustomerCustom() interface{} {
	return &CustomerCustom{}
}

func (cp FullOrderCustomProvider) Fields() *bson.M {
	return nil
}
