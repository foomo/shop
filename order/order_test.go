package order

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func getMockPersistor() *Persistor {
	url := os.Getenv("SHOP_MONGO_TEST_URL")
	if len(url) == 0 {
		panic("please export SHOP_MONGO_TEST_URL=127.0.0.1/foomo-shop-orders")
	}
	p, err := NewPersistor(url)
	if err != nil {
		panic(err)
	}
	p.getCollection().DropCollection()
	return p
}

type TestOrder struct {
	Foo       string
	Bar       int
	EmptyTest string `bson:"ämtü,omitempty" mapstructure:"ämtü"`
	Order     *Order
}

func TestPersistor(t *testing.T) {
	p := getMockPersistor()
	customOrder := &TestOrder{
		Foo:       "foo",
		EmptyTest: "not that empty",
		Bar:       3,
	}
	o := NewOrder(customOrder)
	o.Lala = " lalalala "
	err := p.Create(o)
	if err != nil {
		panic(err)
	}
	loadedOrders, err := p.Find(&bson.M{}, func() interface{} { return &TestOrder{} })
	if err != nil {
		panic(err)
	}
	if len(loadedOrders) != 1 {
		t.Fatal("wrong number of orders returned")
	}

	loadedOrder := loadedOrders[0].Custom.(*TestOrder)
	b, _ := json.MarshalIndent(loadedOrders[0], "", "   ")
	t.Log(reflect.ValueOf(loadedOrder).Type(), string(b))

	if !reflect.DeepEqual(loadedOrder, customOrder) {
		t.Fatal("should have been equal", loadedOrder, customOrder)
	}
	//LoadOrder(query *bson.M{}, customOrder interface{})
}
