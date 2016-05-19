package customer

import (
	"errors"
	"log"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/event_log"
	"github.com/foomo/shop/persistence"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/mgo.v2/bson"
)

// !! NOTE: customer must not import order !!

const (
	VERBOSE = false
)

var globalCustomerPersistor *persistence.Persistor

// NewPersistor constructor
func NewPersistor(mongoURL string, collectionName string) (p *persistence.Persistor, err error) {
	return persistence.NewPersistor(mongoURL, collectionName)
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func GetCustomerPersistor() *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_CUSTOMERS
	if globalCustomerPersistor == nil {
		p, err := NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB global persistor: " + err.Error()))
		}
		globalCustomerPersistor = p
		return globalCustomerPersistor
	}

	if url == globalCustomerPersistor.GetURL() && collection == globalCustomerPersistor.GetCollectionName() {
		return globalCustomerPersistor
	}

	p, err := NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalCustomerPersistor = p
	return globalCustomerPersistor
}

// GetCustomer retrieves a single customer from the database
func GetCustomer(customerId string, customCustomerProvider CustomerCustomProvider) (*Customer, error) {
	iter, err := Find(&bson.M{"id": customerId}, customCustomerProvider)
	if err != nil {
		log.Println(err.Error())
		event_log.SaveShopEvent(event_log.ActionRetrieveCustomer, customerId, err, "")
		return nil, err
	}
	customer, err := iter()
	if err != nil {
		log.Println(err.Error())
		event_log.SaveShopEvent(event_log.ActionRetrieveCustomer, customerId, err, "")
		return nil, err
	}
	if customer == nil {
		log.Println(err.Error())
		event_log.SaveShopEvent(event_log.ActionRetrieveCustomer, customerId, errors.New("No customer with id "+customerId), "")
		return nil, err
	}
	return customer, nil
}

func AlreadyExistsInDB(customerID string) (bool, error) {
	p := GetCustomerPersistor()
	q := p.GetCollection().Find(&bson.M{"id": customerID})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Find returns an iterator for all entries found in database.
func Find(query *bson.M, customProvider CustomerCustomProvider) (iter func() (cust *Customer, err error), err error) {
	p := GetCustomerPersistor()
	_, err = p.GetCollection().Find(query).Count()
	if err != nil {
		log.Println(err)
	}
	//log.Println("Persistor.Find(): ", n, "items found for query ", query)
	q := p.GetCollection().Find(query)
	fields := customProvider.Fields()
	if fields != nil {
		q.Select(fields)
	}
	_, err = q.Count()
	if err != nil {
		return
	}
	mgoiter := q.Iter()
	iter = func() (cust *Customer, err error) {
		cust = &Customer{}

		if mgoiter.Next(cust) {

			/* Map CustomerCustom */
			customerCustom := customProvider.NewCustomerCustom()
			if customerCustom != nil && cust.Custom != nil {
				err = mapstructure.Decode(cust.Custom, customerCustom)
				if err != nil {
					return nil, err
				}
				cust.Custom = customerCustom
			}

			/* Map AddressCustom */
			for _, address := range cust.Addresses {
				addressCustom := customProvider.NewAddressCustom()
				if addressCustom != nil && address.Custom != nil {

					err = mapstructure.Decode(address.Custom, addressCustom)
					if err != nil {
						return nil, err
					}
					address.Custom = addressCustom
				}
			}

			return cust, nil
		}
		return nil, nil
	}

	return
}

func InsertCustomer(c *Customer) error {
	p := GetCustomerPersistor()
	alreadyExists, err := AlreadyExistsInDB(c.GetID())
	if err != nil {
		return err
	}
	if alreadyExists {
		log.Println("User with id", c.GetID(), "already exists in the database!")
		return nil
	}
	err = p.GetCollection().Insert(c)
	event_log.SaveShopEvent(event_log.ActionInsertingCustomer, c.GetID(), err, "")
	return err
}

func UpsertCustomer(c *Customer) error {
	// customer is unlinked or not yet inserted in db
	if c.unlinkDB || c.BsonID == "" {
		return nil
	}
	p := GetCustomerPersistor()
	_, err := p.GetCollection().UpsertId(c.BsonID, c)
	if err != nil {
		panic(err)
	}
	event_log.SaveShopEvent(event_log.ActionUpsertingCustomer, c.GetID(), err, "")
	return err
}
