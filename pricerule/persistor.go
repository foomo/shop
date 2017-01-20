package pricerule

import (
	"errors"
	"reflect"

	"gopkg.in/mgo.v2/bson"

	"github.com/foomo/shop/configuration"
	"github.com/foomo/shop/persistence"
	"github.com/foomo/shop/shop_error"
	"github.com/mitchellh/mapstructure"
)

//Public constants
const (
	TypePriceRules         string = "pricerules"
	TypePriceRulesVouchers string = "vouchers"
	TypePriceRulesGroups   string = "groups"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

var globalPriceRulePersistors map[string]*persistence.Persistor

// PriceRuleCustomProvider - in case obj is extended
type PriceRuleCustomProvider interface {
	NewPriceRuleCustom() interface{}
	NewVoucherCustom() interface{}
	NewGroupCustom() interface{}
	Fields() *bson.M
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// GetGroupByID returns the group with id
func GetGroupByID(ID string, customProvider PriceRuleCustomProvider) (*Group, error) {
	gr, err := findOneObj(new(Group), &bson.M{"id": ID}, nil, "", customProvider)
	if err != nil {
		return nil, err
	}
	return gr.(*Group), err
}

// GetVoucherByID returns the voucher with id
func GetVoucherByID(ID string, customProvider PriceRuleCustomProvider) (*Voucher, error) {
	voucher, err := findOneObj(new(Voucher), &bson.M{"id": ID}, nil, "", customProvider)
	if err != nil {
		return nil, err
	}
	return voucher.(*Voucher), err
}

// GetVoucherByCode returns the voucher with code
func GetVoucherByCode(code string, customProvider PriceRuleCustomProvider) (*Voucher, error) {
	voucher, err := findOneObj(new(Voucher), &bson.M{"vouchercode": code}, nil, "", customProvider)
	if err != nil {
		return nil, err
	}
	return voucher.(*Voucher), err
}

// GetPriceRuleByID returns the group with id
func GetPriceRuleByID(ID string, customProvider PriceRuleCustomProvider) (*PriceRule, error) {
	priceRule, err := findOneObj(new(PriceRule), &bson.M{"id": ID}, nil, "", customProvider)
	if err != nil {
		return nil, err
	}
	return priceRule.(*PriceRule), err
}

// ObjectOfTypeAlreadyExistsInDB checks if a customer with given customerID already exists in the database
func ObjectOfTypeAlreadyExistsInDB(ID string, objOfType interface{}) (bool, error) {
	p := GetPersistorForObject(objOfType)
	q := p.GetCollection().Find(&bson.M{"id": ID})
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetPersistorForObject - fun using reflection/
func GetPersistorForObject(obj interface{}) *persistence.Persistor {
	//val := reflect.ValueOf(obj)
	switch obj.(type) { //attrType.String() {
	case *Group:
		return getPriceRulePersistorForType(TypePriceRulesGroups)
	case *Voucher:
		return getPriceRulePersistorForType(TypePriceRulesVouchers)
	case *PriceRule:
		return getPriceRulePersistorForType(TypePriceRules)
	default:
		attrType := reflect.TypeOf(obj)
		panic("unsupported persistor for type" + attrType.Name())
	}
}

// Returns GLOBAL_PERSISTOR. If GLOBAL_PERSISTOR is nil, a new persistor is created, set as GLOBAL_PERSISTOR and returned
func getPriceRulePersistorForType(persistorType string) *persistence.Persistor {
	url := configuration.MONGO_URL
	collection := configuration.MONGO_COLLECTION_PRICERULES
	switch persistorType {
	case TypePriceRules:
		collection = configuration.MONGO_COLLECTION_PRICERULES

	case TypePriceRulesVouchers:
		collection = configuration.MONGO_COLLECTION_PRICERULES_VOUCHERS

	case TypePriceRulesGroups:
		collection = configuration.MONGO_COLLECTION_PRICERULES_GROUPS

	default:
		panic("type " + persistorType + " does not exist")
	}

	if globalPriceRulePersistors[persistorType] == nil {
		p, err := persistence.NewPersistor(url, collection)
		if err != nil || p == nil {
			panic(errors.New("failed to create mongoDB itemCollection persistor: " + err.Error()))
		}
		if len(globalPriceRulePersistors) == 0 {
			globalPriceRulePersistors = make(map[string]*persistence.Persistor)
		}
		globalPriceRulePersistors[persistorType] = p
		return globalPriceRulePersistors[persistorType]
	}

	if url == globalPriceRulePersistors[persistorType].GetURL() && collection == globalPriceRulePersistors[persistorType].GetCollectionName() {
		return globalPriceRulePersistors[persistorType]
	}

	p, err := persistence.NewPersistor(url, collection)
	if err != nil || p == nil {
		panic(err)
	}
	globalPriceRulePersistors[persistorType] = p
	return globalPriceRulePersistors[persistorType]
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------
// findOneGroup returns one Group from the database
func findOneObj(obj interface{}, find *bson.M, selection *bson.M, sort string, customProvider PriceRuleCustomProvider) (interface{}, error) {
	var p *persistence.Persistor
	p = GetPersistorForObject(obj)

	if find == nil {
		find = &bson.M{}
	}
	if selection == nil {
		selection = &bson.M{}
	}
	if sort != "" {
		err := p.GetCollection().Find(find).Select(selection).Sort(sort).One(obj)
		if err != nil {
			return nil, err
		}
	} else {
		err := p.GetCollection().Find(find).Select(selection).One(obj)
		if err != nil {
			return nil, err
		}
	}

	if customProvider != nil {
		var err error
		typedObject, err := mapDecodeObj(obj, customProvider)
		if err != nil {
			return nil, err
		}
		obj = typedObject
	}

	if obj == nil {
		return nil, errors.New(string(shop_error.ErrorNotFound))
	}
	return obj, nil
}

// mapDecode maps interfaces to specific types provided by customProvider
func mapDecodeObj(obj interface{}, customProvider PriceRuleCustomProvider) (typedObject interface{}, err error) {
	/* Map CustomerCustom */

	switch obj.(type) { //attrType.String() {
	case *Group:
		typedObject := obj.(*Group)
		objCustom := customProvider.NewGroupCustom()
		if objCustom != nil && typedObject.Custom != nil {
			err = mapstructure.Decode(typedObject.Custom, objCustom)
			if err != nil {
				return nil, err
			}
			typedObject.Custom = objCustom
			return typedObject, nil
		}
		return typedObject, nil
	case *Voucher:
		typedObject := obj.(*Voucher)
		objCustom := customProvider.NewVoucherCustom()
		if objCustom != nil && typedObject.Custom != nil {
			err = mapstructure.Decode(typedObject.Custom, objCustom)
			if err != nil {
				return nil, err
			}
			typedObject.Custom = objCustom
			return typedObject, nil
		}
		return typedObject, nil
	case *PriceRule:
		typedObject := obj.(*Voucher)
		objCustom := customProvider.NewPriceRuleCustom()
		if objCustom != nil && typedObject.Custom != nil {
			err = mapstructure.Decode(typedObject.Custom, objCustom)
			if err != nil {
				return nil, err
			}
			typedObject.Custom = objCustom
			return typedObject, nil
		}
		return typedObject, nil
	default:
		return nil, errors.New("unknown type " + reflect.TypeOf(obj).String())
	}

}
