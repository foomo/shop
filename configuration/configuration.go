package configuration

const (
	MONGO_URL                                              string = "mongodb://127.0.0.1/project-globus-services-stories"
	MONGO_COLLECTION_CREDENTIALS                           string = "credentials"
	MONGO_COLLECTION_SHOP_EVENT_LOG                        string = "shop_event_log"
	MONGO_COLLECTION_ORDERS                                string = "orders"
	MONGO_COLLECTION_ORDERS_HISTORY                        string = "orders_history"
	MONGO_COLLECTION_CUSTOMERS_HISTORY                     string = "customers_history"
	MONGO_COLLECTION_CUSTOMERS                             string = "customers"
	MONGO_COLLECTION_MOCK_TRANSACTIONS                     string = "mock_trx"                 // Several TLOGS from globus-service-docs TODO Remove this when no longer used
	MONGO_COLLECTION_MOCK_TRANSACTIONS_FOR_CASHREPORT_TEST string = "mock_trx_cashreport_test" // The one TLOG where we have a corresponding CashReport TODO Remove this when no longer used
	MONGO_COLLECTION_MOCK_ITEMS                            string = "mock_items"               // TODO Remove this when no longer used

)
