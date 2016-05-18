package configuration

const (
	MONGO_URL                                              string = "mongodb://127.0.0.1/project-globus-services-stories"
	MONGO_COLLECTION_SHOP_EVENT_LOG                        string = "shop_event_log"
	MONGO_COLLECTION_ORDERS                                string = "orders"
	MONGO_COLLECTION_CUSTOMERS                             string = "customers"
	MONGO_COLLECTION_MOCK_TRANSACTIONS                     string = "mock_trx"                 // Several TLOGS from globus-service-docs TODO Remove this when no longer used
	MONGO_COLLECTION_MOCK_TRANSACTIONS_FOR_CASHREPORT_TEST string = "mock_trx_cashreport_test" // The one TLOG where we have a corresponding CashReport TODO Remove this when no longer used
	MONGO_COLLECTION_MOCK_ITEMS                            string = "mock_items"               // TODO Remove this when no longer used

	//MONGO_COLLECTION_QUEUE_TEST     string = "queue_test"
	TestSAPServer        = "https://vspid950.gmsap.migros.ch:8444"        //  (goes to localhost and is then tunneled to sap via dev-server)
	TestGAUTServer       = "https://autorisierungstestews.globus.ch:8445" //  (goes to localhost and is then tunneled to GAUT via dev-server)
	TestCadeauCardServer = "https://cadeaucardtestews.globus.ch:8446"
	TestScoringServer    = "https://preprodservices.crif-online.ch"
	TestSendMockRequest  = "https://globus-shop-dev.bestbytes.net:443"
)

var ORDER_ID_RANGE = []int{9200001, 9299999}

//type StaticConfiguration struct {
//	ShippingProvider  string
//	SalesOrganistaion string
//	SalesOffice       string
//	SalesChannel      string
//	CurrencyCode      string
//}
