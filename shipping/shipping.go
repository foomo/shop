package shipping

type ShippingProperties struct {
	ShippingMode     string
	IsSingleShipment bool
	DeliveryStop     string
	//DesiredDeliveryDate *types.XsdDateDay
	Comment string
}
