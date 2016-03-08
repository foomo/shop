package examples_test

import (
	"fmt"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/order"
)

func ExampleOrderCustom_createCart() {
	// a cart is an incomplete order
	o := order.NewOrder()
	o.Custom = &examples.SmurfOrderCustom{
		ResponsibleSmurf: "Pete",
	}
	const (
		positionIDA = "awesome-computer-a"
		positionIDB = "awesome-computer-b"
	)

	// add a product
	o.AddPosition(&order.Position{
		ItemID:   positionIDA,
		Name:     "an awesome computer",
		Quantity: 1.0,
		Custom: &examples.SmurfPositionCustom{
			Foo: "foo",
		},
	})

	// set qty
	if o.SetPositionQuantity(positionIDA, 3.01) != nil {
		panic("could not set qty")
	}

	// add another position
	o.AddPosition(&order.Position{
		ItemID:   positionIDB,
		Name:     "an awesome computer",
		Quantity: 1.0,
		Custom: &examples.SmurfPositionCustom{
			Foo: "bar",
		},
	})

	o.SetPositionQuantity(positionIDB, 0)

	fmt.Println(
		"responsible smurf:",
		o.Custom.(*examples.SmurfOrderCustom).ResponsibleSmurf,
		", position foo:",
		o.Positions[0].Custom.(*examples.SmurfPositionCustom).Foo,
		", qty:",
		o.Positions[0].Quantity,
		", number of positions:",
		len(o.Positions),
	)
	// Output: responsible smurf: Pete , position foo: foo , qty: 3.01 , number of positions: 1
}
