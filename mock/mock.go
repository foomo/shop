package mock

import (
	"fmt"
	"os"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/order"
	"github.com/foomo/shop/queue"
)

const (
	MOCK_EMAIL     = "Foo@Bar.com"
	MOCK_PASSWORD  = "supersafepassword!11"
	MOCK_EMAIL2    = "Alice@Bar.com"
	MOCK_PASSWORD2 = "evensaferpassword!11!§$%&"
)

func MockMongoURL() string {
	url := os.Getenv("SHOP_MONGO_TEST_URL")
	if len(url) == 0 {
		url = "mongodb://127.0.0.1/foomo-shop-orders-mock"
	}
	return url
}

func GetMockQueue() *queue.Queue {
	q, err := queue.NewQueue()
	if err != nil {
		panic(err)
	}
	return q
}

func MakeMockOrder(smurf string) *order.Order {
	custom := &examples.SmurfOrderCustom{
		ResponsibleSmurf: smurf,
	}
	o, err := order.NewOrder(&examples.SmurfOrderCustomProvider{})
	if err != nil {
		panic(err)
	}
	o.Custom = custom
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
