package configuration

const (
	MONGO_DB string = "shop"

	LocalUnitTests = "dockerhost/" + MONGO_DB
	WithDocker     = "mongo/" + MONGO_DB

	//MONGO_URL string = "mongodb://" + LocalUnitTests
	MONGO_URL string = "mongodb://" + WithDocker

	MONGO_COLLECTION_CREDENTIALS       string = "credentials"
	MONGO_COLLECTION_SHOP_EVENT_LOG    string = "shop_event_log"
	MONGO_COLLECTION_ORDERS            string = "orders"
	MONGO_COLLECTION_ORDERS_HISTORY    string = "orders_history"
	MONGO_COLLECTION_CUSTOMERS_HISTORY string = "customers_history"
	MONGO_COLLECTION_CUSTOMERS         string = "customers"
	MONGO_COLLECTION_WATCHLISTS        string = "watchlists"

	MONGO_COLLECTION_PRICERULES          string = "pricerules"
	MONGO_COLLECTION_PRICERULES_VOUCHERS string = "pricerules_vouchers"
	MONGO_COLLECTION_PRICERULES_GROUPS   string = "pricerules_groups"
)

// AllowedLanguages contains language codes for all allowed languages
var AllowedLanguages = [...]string{"de", "fr"}
