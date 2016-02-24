package shipping

import "git.bestbytes.net/Project-Globus-Services/types"

type ShippingProperties struct {
	ShippingMode        string
	IsSingleShipment    bool
	DeliveryStop        string
	DesiredDeliveryDate *types.XsdDateDay
	Comment             string
}
