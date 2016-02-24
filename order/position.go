package order

// Position in an order
type Position struct {
	ID           string
	ItemNumber   string
	Name         string
	Description  string
	Quantity     float64
	QuantityUnit string
	Price        float64
	Custom       interface{}
}
