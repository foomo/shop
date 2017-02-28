package pricerule

import (
	"log"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------
const (
	TypePromotionCustomer     Type = "promotion_customer"      // if applied
	TypePromotionProduct      Type = "promotion_product"       //only one per product can be applied
	TypePromotionOrder        Type = "promotion_order"         // multiple can be applied
	TypeVoucher               Type = "voucher"                 // rule associated to a voucher
	TypePaymentMethodDiscount Type = "payment_method_discount" // rule associated to a payment method

	ActionItemByPercent ActionType = "item_by_percent"
	ActionCartByPercent ActionType = "cart_by_percent"

	ActionCartByAbsolute ActionType = "cart_by_absolute"
	ActionItemByAbsolute ActionType = "item_by_absolute"

	ActionBuyXPayY ActionType = "buy_x_pay_y"
	ActionScaled   ActionType = "scaled"

	XYCheapestFree      XYWhichType = "xy-cheapest-free"
	XYMostExpensiveFree XYWhichType = "xy-most-expensive-free"
)

// MaxUint const
const MaxUint = ^uint(0)

// MinUint const
const MinUint = 0

//MaxInt const
const MaxInt = int(MaxUint >> 1)

//MinInt const
const MinInt = -MaxInt - 1

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

//PriceRule the price rule type
type PriceRule struct {
	ID string //unique id of the price rule

	MappingID string // used to retrieve PromoID and ActionType from Mappings // @todo maybe move to Custom as it's not generic

	Type Type // one of Type, e.g promotion

	Action ActionType //Action - what the price rule does

	Amount float64 //the value depending on action

	Priority int // the articleCollection in which rule is applied

	ValidFrom time.Time // valid from timestamp

	ValidTo time.Time //valid to timestamp

	Exclusive bool // only apply this rule or not, e.g. allow discount accumulation

	Name map[string]string //localized name

	Description map[string]string //localized description

	ExcludedProductGroupIDS []string //ExcludedGroupIds - the voucher group ids for which this rule DOES NOT apply

	IncludedProductGroupIDS []string //IncludedGroupIds - the voucher group ids for which this rule MUST BE PRESSENT to be applicable

	ExcludedCustomerGroupIDS []string //ExcludedCustomerGroupIds -  for which this rule DOES NOT apply

	IncludedCustomerGroupIDS []string //IncludedGroupIds -  MUST BE PRESSENT to be applicable

	CheckoutAttributes []string // for CheckoutAttributes -  payment methods etc - only applicable if the checkout provides these attribues

	X int // buy X - the number of items that are applicable for an ActionBuyXGetY pricerule, for example order 4 pay 3 means X=4 and Y=3

	Y int // get Y - the number of items that one has to pay for for an ActionBuyXGetY price, so X-Y are free, X>=Y

	WhichXYFree XYWhichType

	QtyThreshold float64 // - the total qty in order for the price rule to be applicable. defaults to 0.

	ScaledAmounts []ScaledAmountLevel //defines discount scale 100 -> 2%, 200 -> 3% etc - See ActionScaledPercentage & ActionScaledAbsolute

	ScaledAmountsPerQuantity []ScaledAmountLevel

	MinOrderAmount float64 //minimum amount for discount to be applocable

	MinOrderAmountApplicableItemsOnly bool // must the min amount be calculated only over the applicable items

	MaxUses int //maximum times a pricerule can be applied globally

	MaxUsesPerCustomer int //maximum number of usages per customer

	UsageHistory struct {
		TotalUsages       int            //total times this was applied
		UsagesPerCustomer map[string]int //times a customer used this rule customerId => times
	}

	CreatedAt time.Time //created at

	LastModifiedAt time.Time // updated at

	Custom interface{} `bson:",omitempty"` //make it extensible if needed (included, excluded group IDs)

}

//Type the type of the price rule
type Type string

//ActionType the type of price rule action
type ActionType string

// TypeRuleValidationMsg -
type TypeRuleValidationMsg string

// XYWhichType -
type XYWhichType string

// ScaledAmountLevel -
type ScaledAmountLevel struct {
	IsFromToPrice            bool
	FromValue                float64
	ToValue                  float64
	Amount                   float64 //if percentage, the amount is 0.0 - 100.0
	IsScaledAmountPercentage bool    // is amount a percentage or absolute
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewPriceRule - set defaults
func NewPriceRule(ID string) *PriceRule {
	priceRule := new(PriceRule)
	priceRule.ID = ID
	priceRule.Name = map[string]string{}
	priceRule.Description = map[string]string{}
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 0
	priceRule.MinOrderAmount = 0
	priceRule.QtyThreshold = 0
	priceRule.MaxUses = MaxInt
	priceRule.MaxUsesPerCustomer = MaxInt
	priceRule.Exclusive = false
	priceRule.ExcludedProductGroupIDS = []string{}
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.CheckoutAttributes = []string{}
	priceRule.MaxUsesPerCustomer = MaxInt
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.Priority = 999
	priceRule.ValidFrom = time.Date(1971, time.January, 1, 0, 0, 0, 0, time.UTC)
	priceRule.ValidTo = time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC) // far in the future
	priceRule.WhichXYFree = XYCheapestFree
	return priceRule
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// LoadPriceRule -
func LoadPriceRule(ID string, customProvider PriceRuleCustomProvider) (*PriceRule, error) {
	return GetPriceRuleByID(ID, customProvider)
}

// PriceRuleAlreadyExistsInDB checks if a PriceRule with given ID already exists in the database
func PriceRuleAlreadyExistsInDB(ID string) (bool, error) {
	return ObjectOfTypeAlreadyExistsInDB(ID, new(PriceRule))
}

func (pricerule *PriceRule) Insert() error {
	exists, err := PriceRuleAlreadyExistsInDB(pricerule.ID)
	if err != nil {
		return err
	}
	if exists {
		log.Println("Did not insert Pricerule with id ", pricerule.ID, " ==> duplicate")
		return nil
	}
	return pricerule.Upsert()

}

// Upsert - upsers a PriceRule
// note that if you programmatically manipulate the CreatedAt time, this methd will upsert it
func (pricerule *PriceRule) Upsert() error {
	//set created and modified times
	if pricerule.CreatedAt.IsZero() {
		priceruleFromDb, err := GetPriceRuleByID(pricerule.ID, nil)
		if err != nil || priceruleFromDb == nil {
			pricerule.CreatedAt = time.Now()
		} else {
			pricerule.CreatedAt = priceruleFromDb.CreatedAt
		}
	}
	pricerule.LastModifiedAt = time.Now()

	p := GetPersistorForObject(pricerule)
	_, err := p.GetCollection().Upsert(bson.M{"id": pricerule.ID}, pricerule)

	if err != nil {
		return err
	}
	return nil
}

// UpdatePriceRuleUsageHistoryAtomic - atomicaly update times used and times used per customer if customer id provided
func UpdatePriceRuleUsageHistoryAtomic(ID string, customerID string) error {
	mutex := sync.Mutex{}

	mutex.Lock()
	defer mutex.Unlock()
	priceRule, err := LoadPriceRule(ID, nil)
	if err != nil {
		return err
	}
	return priceRule.UpdateUsageHistory(customerID)

}

// UpdateUsageHistory -
func (pricerule *PriceRule) UpdateUsageHistory(customerID string) error {
	pricerule.UsageHistory.TotalUsages++
	//init map
	if pricerule.UsageHistory.UsagesPerCustomer == nil {
		pricerule.UsageHistory.UsagesPerCustomer = make(map[string]int)
	}
	//if we have a customer
	if len(customerID) > 0 {
		pricerule.UsageHistory.UsagesPerCustomer[customerID]++
	}
	log.Println("updated rule usage history: " + pricerule.ID)
	return pricerule.Upsert()
}

// Delete - delete PriceRule - ID must be set
func (pricerule *PriceRule) Delete() error {
	err := GetPersistorForObject(pricerule).GetCollection().Remove(bson.M{"id": pricerule.ID})
	pricerule = nil
	return err
}

// DeletePriceRule - delete PriceRule
func DeletePriceRule(ID string) error {
	err := GetPersistorForObject(new(PriceRule)).GetCollection().Remove(bson.M{"id": ID})
	return err
}

// RemoveAllPriceRules -
func RemoveAllPriceRules() error {
	p := GetPersistorForObject(new(PriceRule))
	_, err := p.GetCollection().RemoveAll(bson.M{})
	return err
}

// GetValidPriceRulesForCheckoutAttributes - find rule for payment method etc etc
// check ValidFrom, ValidTo
func GetValidPriceRulesForCheckoutAttributes(checkoutAttributes []string, customProvider PriceRuleCustomProvider) ([]PriceRule, error) {

	paymentPriceruleTypes := []Type{TypePromotionCustomer, TypePromotionProduct, TypePromotionOrder, TypePaymentMethodDiscount}
	query := bson.M{"type": bson.M{"$in": paymentPriceruleTypes}, "includedcheckoutattributes": bson.M{"$in": checkoutAttributes}, "validfrom": bson.M{"$lte": time.Now()}, "validto": bson.M{"$gte": time.Now()}}
	return getPromotions(query, customProvider)
}

// GetValidPriceRulesForPromotions - find rule for payment
// check ValidFrom, ValidTo
func GetValidPriceRulesForPromotions(priceRuleTypes []Type, customProvider PriceRuleCustomProvider) ([]PriceRule, error) {
	query := bson.M{"type": bson.M{"$in": priceRuleTypes}, "includedcheckoutattributes": bson.M{"$exists": true, "$size": 0}, "validfrom": bson.M{"$lte": time.Now()}, "validto": bson.M{"$gte": time.Now()}}
	return getPromotions(query, customProvider)
}

func getPromotions(query bson.M, customProvider PriceRuleCustomProvider) ([]PriceRule, error) {
	now := time.Now()
	var result []*PriceRule

	p := GetPersistorForObject(new(PriceRule))
	err := p.GetCollection().Find(query).Select(nil).Sort("priority").All(&result)
	if err != nil {
		// handle error
		return nil, err
	}

	if customProvider == nil {
		priceRulesMapped := []PriceRule{}
		for _, r := range result {
			priceRulesMapped = append(priceRulesMapped, *r)
		}
		return priceRulesMapped, nil
	}

	priceRulesMapped := []PriceRule{}
	for _, r := range result {

		if customProvider != nil {
			var err error
			typedObject, err := mapDecodeObj(r, customProvider)
			if err != nil {
				return nil, err
			}
			r = typedObject.(*PriceRule)
			priceRulesMapped = append(priceRulesMapped, *r)
		}
	}
	timeTrack(now, "[GetValidPriceRulesForPromotions] cache loading took ")
	return priceRulesMapped, nil

}

//------------------------------------------------------------------
// ~ PRICERULE PAIR SORTING - BY PRIORITY - higher Priority means it is applied first
//------------------------------------------------------------------

// ByPriority -
type ByPriority []RuleVoucherPair

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Rule.Priority > a[j].Rule.Priority }
