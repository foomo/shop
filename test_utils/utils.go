package test_utils

import (
	"path"
	"runtime"
	"strings"

	"gopkg.in/mgo.v2"

	"github.com/foomo/shop/configuration"
)

// These collections are not dropped on DropAllCollections
var NoDropList = map[string]bool{
	"erv_invoice_numbers":      true,
	"mock_trx":                 true,
	"mock_trx_cashreport_test": true,
	"orders_many":              true,
	"status_updates_many":      true,
	// The following are all filled by promotions.Initialiue in main Project
	"promo_customer_group":        true,
	"promo_crm":                   true,
	"promo_sap_pricetype_mapping": true,
	"promo_prokey_mapping":        true,
	"promo_actiontype_mapping":    true,
}

func GetTestUtilsDir() string {
	_, filename, _, _ := runtime.Caller(1)
	filename = strings.Replace(filename, "/test_utils.go", "", -1) // remove "utils.go"
	return path.Dir(filename)                                      // remove //"utils" and return
}

// Drops order collection

func DropAllShopCollections() error {
	err := DropAllCollectionsFromUrl(configuration.GetMongoURL(), configuration.MONGO_DB)
	if err != nil {
		return err
	}
	return nil
}

func DropAllCollections() error {
	err := DropAllCollectionsFromUrl(configuration.GetMongoURL(), configuration.MONGO_DB)
	if err != nil {
		return err
	}
	err = DropAllCollectionsFromUrl(configuration.GetMongoProductsURL(), configuration.MONGO_DB_PRODUCTS)
	if err != nil {
		return err
	}
	return nil
}
func DropAllCollectionsFromUrl(url string, db string) error {
	session, err := mgo.Dial(url)
	if err != nil {
		return err
	}
	defer session.Close()


	collections, err := session.DB(db).CollectionNames()

	for _, collectionName := range collections {
		_, ok := NoDropList[collectionName]
		// Only Drop Collections which are not on the no drop list
		if !ok {
			collection := session.DB(db).C(collectionName)
			err = collection.DropCollection()
		}
	}
	return nil
}
