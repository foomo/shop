package configuration

const (
	MONGO_DB string = "project-globus-services-stories"

	LocalUnitTests = "dockerhost/" + MONGO_DB
	WithDocker     = "mongo/project-globus-services-1"

	//MONGO_URL string = "mongodb://" + LocalUnitTests
	MONGO_URL                          string = "mongodb://" + WithDocker
	MONGO_COLLECTION_CREDENTIALS       string = "credentials"
	MONGO_COLLECTION_SHOP_EVENT_LOG    string = "shop_event_log"
	MONGO_COLLECTION_ORDERS            string = "orders"
	MONGO_COLLECTION_ORDERS_HISTORY    string = "orders_history"
	MONGO_COLLECTION_CUSTOMERS_HISTORY string = "customers_history"
	MONGO_COLLECTION_CUSTOMERS         string = "customers"
)
