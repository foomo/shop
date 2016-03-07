package configuration

const (
	MONGO_URL                       string = "mongodb://127.0.0.1/project-globus-services-stories"
	MONGO_COLLECTION_SHOP_EVENT_LOG string = "shop_event_log"
)

type StaticConfiguration struct {
	ShippingProvider  string
	SalesOrganistaion string
	SalesOffice       string
	SalesChannel      string
	CurrencyCode      string
}
