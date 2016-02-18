package order

// Position in an order
type Position struct {
	ID          string
	Price       float64
	Quantity    float64
	Name        string
	Description string
	Custom      interface{}
}
