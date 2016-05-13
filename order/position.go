package order

// Position in an order
type Position struct {
	//ID           string
	ItemID       string
	Name         string
	Description  string
	Quantity     float64
	QuantityUnit string
	Price        float64
	IsATPApplied bool
	Refund       bool
	Custom       interface{}
}

func (p *Position) IsRefund() bool {
	return p.Refund
}

// GetAmount returns the Price Sum of the position
func (p *Position) GetAmount() float64 {
	return p.Price * p.Quantity
}
