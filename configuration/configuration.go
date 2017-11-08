package configuration

import "os"

var (
	MONGO_DB          = "shop"
	MONGO_DB_PRODUCTS = "products"
	LocalUnitTests    = "localhost/" + MONGO_DB

	WithDocker = "mongo/" + MONGO_DB

	ProductsLocal  = "localhost/" + MONGO_DB_PRODUCTS
	ProductsDocker = "mongo/" + MONGO_DB_PRODUCTS

	// MONGO_BASE_URL = "mongodb://dockerhost/"
	// //MONGO_BASE_URL = "mongodb://mongo/"

	//MONGO_URL string = "mongodb://" + LocalUnitTests
	MONGO_URL          = "mongodb://" + WithDocker
	MONGO_URL_PRODUCTS = "mongodb://" + ProductsLocal

	MONGO_COLLECTION_CREDENTIALS       = "credentials"
	MONGO_COLLECTION_ORDERS            = "orders"
	MONGO_COLLECTION_ORDERS_HISTORY    = "orders_history"
	MONGO_COLLECTION_CUSTOMERS_HISTORY = "customers_history"
	MONGO_COLLECTION_CUSTOMERS         = "customers"
	MONGO_COLLECTION_WATCHLISTS        = "watchlists"

	MONGO_COLLECTION_PRICERULES          = "pricerules"
	MONGO_COLLECTION_PRICERULES_VOUCHERS = "pricerules_vouchers"
	MONGO_COLLECTION_PRICERULES_GROUPS   = "pricerules_groups"
)

// AllowedLanguages contains language codes for all allowed languages
var AllowedLanguages = [...]string{"de", "fr"}

func GetMongoURL() string {
	if url, exists := os.LookupEnv("MONGO_URL"); exists {
		return url
	}
	return MONGO_URL
}

func GetMongoProductsURL() string {
	if url, exists := os.LookupEnv("MONGO_URL_PRODUCTS"); exists {
		return url
	}
	return MONGO_URL_PRODUCTS
}
