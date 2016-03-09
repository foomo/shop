package mock

import (
	"fmt"
	"os"

	"github.com/foomo/shop/examples"
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

func GetMockQueue() *queue.Queue {
	q, err := queue.NewQueue(MockMongoURL())
	if err != nil {
		panic(err)
	}
	return q
}

func MakeMockOrder(smurf string) *order.Order {
	o := order.NewOrder()
	o.Custom = &examples.SmurfOrderCustom{
		ResponsibleSmurf: smurf,
	}
	for i := 0; i < 5; i++ {
		// add a product
		o.AddPosition(&order.Position{
			ItemID:   "asdf",
			Name:     fmt.Sprintf("an awesome computer - %d", i),
			Quantity: float64(i),
			Custom: &examples.SmurfPositionCustom{
				Foo: fmt.Sprintf("foo - %d", i),
			},
		})

	}
	return o
}
