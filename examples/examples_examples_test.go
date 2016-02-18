package examples_test

import (
	"fmt"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/order"
)

func ExampleOrderCustom() {
	order := order.NewOrder(&examples.OrderCustom{
		ResponsibleSmurf: "Pete",
	})
	fmt.Println(order.Custom.(*examples.OrderCustom).ResponsibleSmurf)
	// Output: Pete
}
