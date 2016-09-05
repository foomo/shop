package configuration

const (
	MONGO_DB      string = "shop"
	MONGO_DB_TEST string = "project-globus-services-stories"

	LocalUnitTests = "dockerhost/" + MONGO_DB_TEST
	WithDocker     = "mongo/" + MONGO_DB

	//MONGO_URL string = "mongodb://" + LocalUnitTests
	MONGO_URL string = "mongodb://" + WithDocker

	MONGO_COLLECTION_CREDENTIALS       string = "credentials"
	MONGO_COLLECTION_SHOP_EVENT_LOG    string = "shop_event_log"
	MONGO_COLLECTION_ORDERS            string = "orders"
	MONGO_COLLECTION_ORDERS_HISTORY    string = "orders_history"
	MONGO_COLLECTION_CUSTOMERS_HISTORY string = "customers_history"
	MONGO_COLLECTION_CUSTOMERS         string = "customers"
)

// AllowedLanguages contains language codes for all allowed languages
var AllowedLanguages = [...]string{"de", "fr"}
