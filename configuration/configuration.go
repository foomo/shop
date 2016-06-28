package configuration

const (
	MONGO_DB                           string = "project-globus-services-stories"
	MONGO_URL                          string = "mongodb://127.0.0.1/project-globus-services-stories"
	MONGO_COLLECTION_CREDENTIALS       string = "credentials"
	MONGO_COLLECTION_SHOP_EVENT_LOG    string = "shop_event_log"
	MONGO_COLLECTION_ORDERS            string = "orders"
	MONGO_COLLECTION_ORDERS_HISTORY    string = "orders_history"
	MONGO_COLLECTION_CUSTOMERS_HISTORY string = "customers_history"
	MONGO_COLLECTION_CUSTOMERS         string = "customers" // TODO Remove this when no longer used
)
