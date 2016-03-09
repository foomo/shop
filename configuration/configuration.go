package configuration

const (
	MONGO_URL                       string = "mongodb://127.0.0.1/project-globus-services-stories"
	MONGO_COLLECTION_SHOP_EVENT_LOG string = "shop_event_log"
	MONGO_COLLECTION_ORDERS         string = "orders"
	//MONGO_COLLECTION_QUEUE_TEST     string = "queue_test"
	TestSAPServer = "https://vspid950.gmsap.migros.ch:8444" //  (goes to localhost and is then tunneled to sap via dev-server)
	//TestServer = "https://127.0.0.1:8444"
	TestSendMockRequest = "https://globus-shop-dev.bestbytes.net:443"
)

var ORDER_ID_RANGE = []int{9200001, 9299999}

type StaticConfiguration struct {
	ShippingProvider  string
	SalesOrganistaion string
	SalesOffice       string
	SalesChannel      string
	CurrencyCode      string
}
