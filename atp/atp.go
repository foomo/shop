package shop

import (
	"time"

	"github.com/foomo/shop/order"
)

type ATPRequest struct {
	Items []*RequestItem `xml:"ITEM,omitempty"`
}

type ATPResponse struct {
	Items []*ResponseItem `xml:"ITEM,omitempty"`
}

type RequestItem struct {
	AssociatedOrder     *order.Order // this is required to get access to custom data when the request is transformed to a Custom Request (e.e. MZG)
	ItemNumber          string
	DesiredQuantity     float64
	QuantityUnit        string
	DesiredDeliveryDate time.Time
	ShippingProvider    string
	Custom              interface{}
}

type ResponseItem struct {
	ItemNumber          string
	ApprovedQuantity    float64
	QuantityUnit        string
	DesiredDeliveryDate time.Time
	DeliveryDate        time.Time
	ShippingProvider    string
	ErrorCode           string
	Custom              interface{}
}

func CreateATPRequestFromOrder(order *order.Order) *ATPRequest {

	requestItems := []*RequestItem{}
	for _, position := range order.Positions {

		reqItem := &RequestItem{
			AssociatedOrder:     order,
			ItemNumber:          position.ItemID,
			DesiredQuantity:     position.Quantity,
			QuantityUnit:        position.QuantityUnit,
			DesiredDeliveryDate: order.Timestamp.Add(time.Duration(1) * time.Hour * time.Duration(24)), // set day after orderdate as default for DesiredDeliveryDate
		}
		requestItems = append(requestItems, reqItem)
	}

	return &ATPRequest{
		Items: requestItems,
	}

}
