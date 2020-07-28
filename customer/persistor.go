package customer

import (
	stderr "errors"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/persistence"
	"github.com/foomo/shop/shop_error"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// !! NOTE: customer must not import order !!
//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var (
	globalCustomerPersistor    *persistence.Persistor
	globalCredentialsPersistor *persistence.Persistor

	customerEnsuredIndexes = []mgo.Index{
		{
			Name:       "addrkey",
			Key:        []string{KeyAddrKey},
			Unique:     true,
			Background: true,
		},
		{
			Name:       "addrkeyhash",
			Key:        []string{KeyAddrKeyHash},
			Unique:     true,
			Background: true,
		},
		mgo.Index{
			Name:       "id",
			Key:        []string{"id"},
			Unique:     true,
			Background: true,
		},
		mgo.Index{
			Name:       "mail",
			Key:        []string{"email"},
			Unique:     false,
			Background: true,
		},
		mgo.Index{
			Name:       "externalid",
			Key:        []string{"externalid"},
			Unique:     true,
			Background: true,
		},
	}
)

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// GetCustomerPersistor will return a singleton instance of a customer mongo persistor
func GetCustomerPersistor() *persistence.Persistor {
	url := configuration.GetMongoURL()
	collection := configuration.MONGO_COLLECTION_CUSTOMERS
	if globalCustomerPersistor == nil {
		p, err := persistence.NewPersistorWithIndexes(url, collection, customerEnsuredIndexes)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB global persistor: " + err.Error()))
		}
		globalCustomerPersistor = p
		return globalCustomerPersistor
	}

	if url == globalCustomerPersistor.GetURL() && collection == globalCustomerPersistor.GetCollectionName() {
		return globalCustomerPersistor
	}

	p, err := persistence.NewPersistorWithIndexes(url, collection, customerEnsuredIndexes)
	if err != nil || p == nil {
		panic(err)
	}
	globalCustomerPersistor = p
	return globalCustomerPersistor
}

// AlreadyExistsInDB checks if a customer with given customerID already exists in the database
func AlreadyExistsInDB(customerID string) (bool, error) {
	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()

	q := collection.Find(&bson.M{"id": customerID})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count will count the items in mongo collection matching the query
func Count(query *bson.M, customProvider CustomerCustomProvider) (count int, err error) {
	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()

	return collection.Find(query).Count()
}

// Find returns an iterator for all entries found matching on query.
func Find(query *bson.M, customProvider CustomerCustomProvider) (iter func() (cust *Customer, err error), err error) {
	q := GetCustomerPersistor().GetGlobalSessionCollection().Find(query)

	_, errCount := q.Count()
	if errCount != nil {
		if stderr.Is(errCount, mgo.ErrNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, errCount
	}

	mgoiter := q.Iter()
	iter = func() (cust *Customer, err error) {
		cust = &Customer{}
		if mgoiter.Next(cust) {
			return MapDecode(cust, customProvider)
		}
		return nil, mgoiter.Err()
	}
	return
}

// UpsertCustomer will save a given customer in mongo collection
func UpsertCustomer(c *Customer) error {

	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()

	if c.Version == nil {
		return errors.Wrap(shop_error.ErrorVersionConflict, "version must not be empty")
	}

	currentVersion := c.Version.Current
	c.Version.Increment()
	if currentVersion == 0 {
		// it is not allowed to insert customers, @see NewCustomer method instead
		return shop_error.ErrorVersionConflict
	}

	// upsert existing customer
	err := collection.Update(bson.M{
		"$and": []bson.M{
			{KeyAddrKey: c.AddrKey},
			{"version.current": currentVersion},
		}}, c)
	if stderr.Is(err, mgo.ErrNotFound) {
		return shop_error.ErrorVersionConflict
	}
	return err
}

func UpsertAndGetCustomer(c *Customer, customProvider CustomerCustomProvider) (*Customer, error) {
	err := UpsertCustomer(c)
	if err != nil {
		return nil, err
	}
	return GetCustomerById(c.GetID(), customProvider)
}

func DeleteCustomer(c *Customer) error {
	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()

	// remove customer
	err := collection.Remove(bson.M{"_id": c.BsonId})
	if stderr.Is(mgo.ErrNotFound, err) {
		return nil
	}
	return err
}
func DeleteCustomerById(id string) error {
	customer, err := GetCustomerById(id, nil)
	if err != nil {
		return err
	}
	DeleteCustomer(customer)
	return err
}

func GetCustomerByAddrKey(addrKey string, customProvider CustomerCustomProvider) (*Customer, error) {
	return GetCustomerByQuery(&bson.M{KeyAddrKey: addrKey}, customProvider)
}

func GetCustomerByAddrKeyHash(addrKeyHash string, customProvider CustomerCustomProvider) (*Customer, error) {
	return GetCustomerByQuery(&bson.M{KeyAddrKeyHash: addrKeyHash}, customProvider)
}

func GetCustomerByQuery(query *bson.M, customProvider CustomerCustomProvider) (*Customer, error) {
	return findOneCustomer(query, nil, "", customProvider)
}

// GetCustomerById returns the customer with id
func GetCustomerById(id string, customProvider CustomerCustomProvider) (*Customer, error) {
	return findOneCustomer(&bson.M{"id": id}, nil, "", customProvider)
}

func DropAllCustomers() error {
	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()

	return collection.DropCollection()

}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// findOneCustomer returns one Customer from the customer database or from the customer history database
func findOneCustomer(find *bson.M, selection *bson.M, sort string, customProvider CustomerCustomProvider) (*Customer, error) {

	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()

	customer := &Customer{}
	if find == nil {
		find = &bson.M{}
	}
	if selection == nil {
		selection = &bson.M{}
	}
	if sort != "" {
		err := collection.Find(find).Select(selection).Sort(sort).One(customer)
		if err != nil {
			if stderr.Is(err, mgo.ErrNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, err
		}
	} else {
		err := collection.Find(find).Select(selection).One(customer)
		if err != nil {
			if stderr.Is(err, mgo.ErrNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, err
		}
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}
	if customProvider != nil {
		var err error
		customer, err = MapDecode(customer, customProvider)
		if err != nil {
			return nil, err
		}
	}

	return customer, nil
}

// insertCustomer inserts a customer into the database
func insertCustomer(c *Customer) error {
	session, collection := GetCustomerPersistor().GetCollection()
	defer session.Close()
	err := collection.Insert(c)
	if mgo.IsDup(err) {
		return errors.Wrap(shop_error.ErrorDuplicateKey, err.Error())
	}
	return err
}

// MapDecode maps interfaces to specific types provided by customProvider
func MapDecode(cust *Customer, customProvider CustomerCustomProvider) (customer *Customer, err error) {
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
