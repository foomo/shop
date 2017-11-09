//Package examples hosts some common examples
package examples

import (
	"time"

	"github.com/foomo/shop/order"
	"github.com/foomo/shop/queue"

	"gopkg.in/mgo.v2/bson"
	"strconv"
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
	query          *bson.M
	Smurf          string
	CountProcessed int
	chanCount      chan int
}

//------------------------------------------------------------------
// ~ CONSTANTS & VARS
//------------------------------------------------------------------

var processorIdSmurf int = 0

//------------------------------------------------------------------
// ~ CONSTRUCTORS
//------------------------------------------------------------------
func NewSmurfProcessor() *queue.DefaultProcessor {
	name := "SmurfProcessor " + strconv.Itoa(processorIdSmurf)
	processorIdSmurf++
	proc := queue.NewDefaultProcessor(name)
	proc.ProcessingFunc = processingFunc
	proc.GetDataWrapper = newOrder
	proc.Persistor = order.GetOrderPersistor()

	return proc
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func newOrder() interface{} {
	return &order.Order{}
}

func processingFunc(v interface{}) error {
	//data, ok := v.(*order.Order)
	// DO SOMETHING WITH data
	time.Sleep(time.Millisecond * 20)
	return nil
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

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
