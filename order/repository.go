package order

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
)

// OverrideID may be used to use a different than the automatially generated id (Unit tests)
func OverrideId(oldID, newID string) error {
	log.Println("+++ INFO: Overriding orderID", oldID, "with id", newID, "+++")
	o := &Order{}
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()

	err := collection.Find(&bson.M{"id": oldID}).One(o)
	if err != nil {
		log.Println("Upsert failed: Could not find order with id", oldID, "Error:", err)
		return err
	}
	o.Id = newID
	o.State.SetModified()
	_, err = collection.UpsertId(o.BsonId, o)
	return err
}

// AlreadyExistsInDB checks if a order with given orderID already exists in the database
func AlreadyExistsInDB(orderID string) (bool, error) {
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()
	q := collection.Find(&bson.M{"id": orderID})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func Count(query *bson.M, customProvider OrderCustomProvider) (count int, err error) {
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()
	return collection.Find(query).Count()
}

// Find returns an iterator for the entries matching on query
func Find(query *bson.M, customProvider OrderCustomProvider) (iter func() (o *Order, err error), err error) {
	collection := GetOrderPersistor().GetGlobalSessionCollection()

	_, err = collection.Find(query).Count()
	if err != nil {
		log.Println(err)
	}
	//log.Println("Persistor.Find(): ", n, "items found for query ", query)
	q := collection.Find(query).Sort("_id")
	// fields := customProvider.Fields()
	// if fields != nil {
	// 	q.Select(fields)
	// }
	_, err = q.Count()
	if err != nil {
		return
	}
	mgoiter := q.Iter()
	iter = func() (o *Order, err error) {
		o = &Order{}

		if mgoiter.Next(o) {
			if customProvider != nil {
				return mapDecode(o, customProvider)
			}
			return o, nil
		}
		return nil, nil
	}
	return
}

func SnapshotByOrderID(orderID string, customProvider OrderCustomProvider) error {
	order, err := GetOrderById(orderID, customProvider)
	if err != nil {
		return err
	}
	return SnapShot(order)
}

func SnapShot(o *Order) error {
	return storeOrderVersionInHistory(o)
}

func UpsertOrder(o *Order) error {

	if o.unlinkDB || o.BsonId == "" {
		return nil
	}
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()

	// Get current version from db and check against version of o
	// If they are not identical, this upsert would be a dirty write
	// In this case upsert is not performed and an error is returned,
	orderLatestFromDb := &Order{}
	err := collection.Find(&bson.M{"id": o.GetID()}).Select(&bson.M{"version": 1}).One(orderLatestFromDb)
	if err != nil {
		return err
	}

	latestVersionInDb := orderLatestFromDb.Version.GetVersion()
	if latestVersionInDb != o.Version.GetVersion() && !o.Flags.forceUpsert {
		errMsg := fmt.Sprintln("WARNING: Cannot upsert latest version ", strconv.Itoa(latestVersionInDb), "in db with version", strconv.Itoa(o.Version.GetVersion()), "!")
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	if o.Flags.forceUpsert {
		// Remember this number, so that we later know from which version we came from
		v := o.Version.Current
		// Set the current version number to keep history consistent
		o.Version.Current = latestVersionInDb
		o.Version.Increment()
		o.Flags.forceUpsert = false
		// Overwrite NumberPrevious, to remember where we came from
		o.Version.Previous = v
	} else {
		o.Version.Increment()
	}

	o.State.SetModified() // this is actually false, but some Globus queries rely on it
	_, err = collection.UpsertId(o.BsonId, o)
	if err != nil {
		return err
	}

	return nil
}

func storeOrderVersionInHistory(o *Order) error {
	currentID := o.BsonId
	o.BsonId = "" // Temporarily reset Mongo ObjectId, so that we can perfrom an Insert.
	session, collection := GetOrderVersionsPersistor().GetCollection()
	defer session.Close()

	err := collection.Insert(o)
	o.BsonId = currentID
	return err
}

func UpsertAndGetOrder(o *Order, customProvider OrderCustomProvider) (*Order, error) {
	err := UpsertOrder(o)
	if err != nil {
		return nil, err
	}
	return GetOrderById(o.GetID(), customProvider)
}

func DeleteOrder(o *Order) error {
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()
	err := collection.Remove(bson.M{"_id": o.BsonId})

	return err
}
func DeleteOrderById(id string) error {
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()
	err := collection.Remove(bson.M{"id": id})
	return err
}

func DropAllOrders() error {
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()

	return collection.DropCollection()

}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// findOneOrder returns one Order from the order database or from the order history database
func findOneOrder(find *bson.M, selection *bson.M, sort string, customProvider OrderCustomProvider, fromHistory bool) (*Order, error) {
	var session *mgo.Session
	var collection *mgo.Collection

	if fromHistory {
		session, collection = GetOrderVersionsPersistor().GetCollection()
	} else {
		session, collection = GetOrderPersistor().GetCollection()
	}
	defer session.Close()

	order := &Order{}
	if find == nil {
		find = &bson.M{}
	}
	if selection == nil {
		selection = &bson.M{}
	}
	if sort != "" {
		err := collection.Find(find).Select(selection).Sort(sort).One(order)
		if err != nil {
			return nil, err
		}
	} else {
		err := collection.Find(find).Select(selection).One(order)
		if err != nil {
			return nil, err
		}
	}
	if customProvider != nil {
		var err error
		order, err = mapDecode(order, customProvider)
		if err != nil {
			return nil, err
		}
	}
	return order, nil
}

// insertOrder inserts a order into the database
func insertOrder(o *Order) error {
	session, collection := GetOrderPersistor().GetCollection()
	defer session.Close()

	alreadyExists, err := AlreadyExistsInDB(o.GetID())
	if err != nil {
		return err
	}
	if alreadyExists {
		log.Println("User with id", o.GetID(), "already exists in the database!")
		return nil
	}
	err = collection.Insert(o)
	if err != nil {
		return err
	}

	hsession, hcollection := GetOrderVersionsPersistor().GetCollection()
	defer hsession.Close()

	err = hcollection.Insert(o)
	return err
}
