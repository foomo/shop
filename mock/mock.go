package mock

import (
	"fmt"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/order"
)


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
