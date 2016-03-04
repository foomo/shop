package shop

import (
	"encoding/xml"

	"time"

	"github.com/foomo/shop/order"
)

type ATPRequest struct {
	XMLName xml.Name       `xml:"sap-skip:http://globus.ch/xi/webshop/atp MT_ATP_REQUEST"`
	Items   []*RequestItem `xml:"ITEM,omitempty"`
}

type ATPResponse struct {
	XMLName xml.Name        `xml:"http://globus.ch/xi/webshop/atp MT_ATP_RESPONSE"`
	Items   []*ResponseItem `xml:"ITEM,omitempty"`
}

type RequestItem struct {
	XMLName             xml.Name     `xml:"ITEM"`
	AssociatedOrder     *order.Order // this is required to get access to custom data when the request is transformed to a Custom Request (e.e. MZG)
	ItemNumber          string       `xml:"artikelnummer,attr,omitempty" validate:"nonzero"`
	DesiredQuantity     float64      `xml:"wunschmenge,attr,omitempty" validate:"nonzero""`
	QuantityUnit        string       `xml:"mengeneinheit,attr,omitempty" "`
	DesiredDeliveryDate time.Time    `xml:"wunschlieferdatum,attr,omitempty"`
	ShippingProvider    string       `xml:"lieferbetrieb,attr,omitempty" `
	Custom              interface{}
}

type ResponseItem struct {
	XMLName             xml.Name  `xml:"ITEM"`
	ItemNumber          string    `xml:"artikelnummer,attr,omitempty" validate:"nonzero"`
	ApprovedQuantity    float64   `xml:"bestaetigte_menge,attr,omitempty" `
	QuantityUnit        string    `xml:"mengeneinheit,attr,omitempty" validate:"max=3"`
	DesiredDeliveryDate time.Time `xml:"wunschlieferdatum,attr,omitempty"`
	DeliveryDate        time.Time `xml:"lieferdatum_1,attr,omitempty" validate:"nonzero"`
	ShippingProvider    string    `xml:"lieferbetrieb,attr,omitempty" validate:”min=1, max=4"` // required
	ErrorCode           string    `xml:"fehlercode,attr,omitempty" `                           // not required
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
