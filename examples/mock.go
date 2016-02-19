package examples

import (
	"fmt"
	"os"

	"github.com/foomo/shop/order"
	"github.com/foomo/shop/queue"
)

func MockMongoURL() string {
	url := os.Getenv("SHOP_MONGO_TEST_URL")
	if len(url) == 0 {
		//panic("please export SHOP_MONGO_TEST_URL=mongodb://127.0.0.1/foomo-shop-orders")
		url = "mongodb://127.0.0.1/foomo-shop-orders-mock"
	}
	return url
}

func GetMockPersistor(collectionName string) *order.Persistor {
	p, err := order.NewPersistor(MockMongoURL(), collectionName)
	if err != nil {
		panic(err)
	}
	p.GetCollection().DropCollection()
	return p
}

func GetMockQueue() *queue.Queue {
	q, err := queue.NewQueue(MockMongoURL())
	if err != nil {
		panic(err)
	}
	return q
}

func MakeMockOrder(smurf string) *order.Order {
	o := order.NewOrder(&OrderCustom{
		ResponsibleSmurf: smurf,
	})
	for i := 0; i < 5; i++ {
		// add a product
		o.AddPosition(&order.Position{
			ID:       "id-" + fmt.Sprint(i),
			Name:     fmt.Sprintf("an awesome computer - %d", i),
			Quantity: float64(i),
			Custom: &PositionCustom{
				Foo: fmt.Sprintf("foo - %d", i),
			},
		})

	}
	return o
}
