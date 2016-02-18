# Cart

A cart is an order, that was not placed.

```go
// a cart is an incomplete order
o := order.NewOrder(&examples.OrderCustom{
    ResponsibleSmurf: "Pete",
})
const (
    positionIDA = "awesome-computer-a"
    positionIDB = "awesome-computer-b"
)

// add a product
o.AddPosition(&order.Position{
    ID:       positionIDA,
    Name:     "an awesome computer",
    Quantity: 1.0,
    Custom: &examples.PositionCustom{
        Foo: "foo",
    },
})
// set qty
if o.SetPositionQuantity(positionIDA, 3.01) != nil {
    panic("could not set qty")
}

// add another position
o.AddPosition(&order.Position{
    ID:       positionIDB,
    Name:     "an awesome computer",
    Quantity: 1.0,
    Custom: &examples.PositionCustom{
        Foo: "bar",
    },
})

o.SetPositionQuantity(positionIDB, 0)

fmt.Println(
    "responsible smurf:",
    o.Custom.(*examples.OrderCustom).ResponsibleSmurf,
    ", position foo:",
    o.Positions[0].Custom.(*examples.PositionCustom).Foo,
    ", qty:",
    o.Positions[0].Quantity,
    ", number of positions:",
    len(o.Positions),
)
// Output: responsible smurf: Pete , position foo: foo , qty: 3.01 , number of positions: 1

```
