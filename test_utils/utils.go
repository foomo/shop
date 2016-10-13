package test_utils

import (
	"log"
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
}

func GetTestUtilsDir() string {
	_, filename, _, _ := runtime.Caller(1)
	filename = strings.Replace(filename, "/test_utils.go", "", -1) // remove "utils.go"
	return path.Dir(filename)                                      // remove //"utils" and return
}

// Drops order collection and event_log collection

func DropAllShopCollections() error {
	err := DropAllCollectionsFromUrl(configuration.MONGO_URL, configuration.MONGO_DB)
	if err != nil {
		return err
	}
	return nil
}
func DropAllCollections() error {
	err := DropAllCollectionsFromUrl(configuration.MONGO_URL, configuration.MONGO_DB)
	if err != nil {
		return err
	}
	err = DropAllCollectionsFromUrl(configuration.MONGO_URL_PRODUCTS, configuration.MONGO_DB_PRODUCTS)
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

	log.Println("mongo session initialized", url, session)

	collections, err := session.DB(db).CollectionNames()
	if err != nil {
		log.Println("unable to find CollectionNames", session)
	} else {
		log.Println("collections", collections)
	}

	for _, collectionName := range collections {
		_, ok := NoDropList[collectionName]
		// Only Drop Collections which are not on the no drop list
		if !ok {
			collection := session.DB(db).C(collectionName)
			count, err := collection.Count()

			if err != nil {
				log.Println("failed to find docs:", collectionName)
			}

			err = collection.DropCollection()
			if err != nil {
				log.Println("failed to drop collection:", collectionName, collection)
			} else {
				log.Printf("dropped collection %s with %d docs", collectionName, count)
			}
		}
	}

	log.Println("DropAllCollections finished")
	return nil
}
