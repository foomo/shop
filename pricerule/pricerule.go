package pricerule

import (
	"errors"
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
	TypeShipping              Type = "shipping"
	TypeBonusVoucher          Type = "bonus_voucher" // rule used to pay with

	ActionItemByPercent ActionType = "item_by_percent"
	ActionCartByPercent ActionType = "cart_by_percent"

	ActionCartByAbsolute ActionType = "cart_by_absolute"
	ActionItemByAbsolute ActionType = "item_by_absolute"

	ActionBuyXPayY ActionType = "buy_x_pay_y"
	ActionScaled   ActionType = "scaled"

	ActionItemSetAbsolute ActionType = "itemset_absolute"

	XYCheapestFree      XYWhichType = "xy-cheapest-free"
	XYMostExpensiveFree XYWhichType = "xy-most-expensive-free"
)

const Verbose = false

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

	IsAmountIndependentOfQty bool // do we apply the discount as amount * qty // false by default

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

	ItemSets [][]string // array of array of skus that represent an item

	X int // buy X - the number of items that are applicable for an ActionBuyXGetY pricerule, for example order 4 pay 3 means X=4 and Y=3

	Y int // get Y - the number of items that one has to pay for for an ActionBuyXGetY price, so X-Y are free, X>=Y

	WhichXYFree XYWhichType

	WhichXYList []string // an ordered list of itemIDs (skus) that affect which item should be given for free. If array not empty WhichXYFree is disregarded

	QtyThreshold float64 // - the total qty in order for the price rule to be applicable. defaults to 0.

	ScaledAmounts []ScaledAmountLevel //defines discount scale 100 -> 2%, 200 -> 3% etc - See ActionScaledPercentage & ActionScaledAbsolute

	ScaledAmountsPerQuantity []ScaledAmountLevel //@todo: remove- but double check if not used somewhere!!!

	MinOrderAmount float64 //minimum amount for discount to be applocable

	MinOrderAmountApplicableItemsOnly bool // must the min amount be calculated only over the applicable items

	CalculateDiscountedOrderAmount bool // shall we use net prices without discount in order total calc(excl tax from item price)

	ExcludedItemIDsFromOrderAmountCalculation []string //items that are not summed up ay order amount calculation

	MaxUses int //maximum times a pricerule can be applied globally

	MaxUsesPerCustomer int //maximum number of usages per customer

	UsageHistory struct {
		TotalUsages       int            //total times this was applied
		UsagesPerCustomer map[string]int //times a customer used this rule customerId => times
	}

	CreatedAt time.Time //created at

	LastModifiedAt time.Time // updated at

	Custom interface{} `bson:",omitempty"` //make it extensible if needed (included, excluded group IDs)

	ExcludeAlreadyDiscountedItemsForVoucher bool

	ExcludesEmployeesForVoucher bool // Note: exclusion of employees must actually be configured by setting IncludedCustomerGroupIDS/ExcludedCustomerGroupIDS.
	// This flag is used for external validation purposes and only provides the information that this promo is supposed to exclude employees.
	// It has no effect on the promo calculation itself!
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
	priceRule.IsAmountIndependentOfQty = false
	priceRule.MinOrderAmount = 0
	priceRule.CalculateDiscountedOrderAmount = false
	priceRule.ExcludedItemIDsFromOrderAmountCalculation = []string{}
	priceRule.QtyThreshold = 0
	priceRule.MaxUses = MaxInt
	priceRule.MaxUsesPerCustomer = MaxInt
	priceRule.Exclusive = false
	priceRule.ExcludedProductGroupIDS = []string{}
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.CheckoutAttributes = []string{}
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.Priority = 999
	priceRule.ValidFrom = time.Date(1971, time.January, 1, 0, 0, 0, 0, time.UTC)
	priceRule.ValidTo = time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC) // far in the future
	priceRule.WhichXYFree = XYCheapestFree
	priceRule.ItemSets = [][]string{}

	return priceRule
}

func NewBonusPriceRule(ruleID string, amount float64, name map[string]string, description map[string]string, validFrom time.Time, validTo time.Time) (priceRule *PriceRule) {
	priceRule = new(PriceRule)
	priceRule.ID = ruleID
	priceRule.Type = TypeBonusVoucher
	priceRule.Name = name
	priceRule.Description = description
	priceRule.Action = ActionCartByAbsolute
	priceRule.Amount = amount
	priceRule.IsAmountIndependentOfQty = false
	priceRule.MinOrderAmount = 0
	priceRule.ExcludedItemIDsFromOrderAmountCalculation = []string{}
	priceRule.QtyThreshold = 0
	priceRule.MaxUses = MaxInt
	priceRule.MaxUsesPerCustomer = 1
	priceRule.Exclusive = false
	priceRule.ExcludedProductGroupIDS = []string{}
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.CheckoutAttributes = []string{}
	priceRule.CalculateDiscountedOrderAmount = true

	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.Priority = 999
	priceRule.ValidFrom = validFrom
	priceRule.ValidTo = validTo
	priceRule.WhichXYFree = XYCheapestFree
	priceRule.ItemSets = [][]string{}
	return
}

func NewBonusVoucher(ruleID string, customerID string, voucherCode string, voucherID string) (voucher *Voucher, err error) {
	if customerID == "" {
		err = errors.New("customer ID must be provided in bonus vouchers")
		return
	}
	voucher = NewVoucherWithRuleID(voucherID, voucherCode, ruleID, customerID)
	return
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
		if Verbose {
			log.Println("Did not insert Pricerule with id ", pricerule.ID, " ==> duplicate")
		}
		return nil
	}
	return pricerule.Upsert()

}

// Upsert - upsers a PriceRule
// note that if you programmatically manipulate the CreatedAt time, this methd will upsert it
func (pricerule *PriceRule) Upsert() error {

	pricerule.checkIfBonusVoucher()

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

	session, collection := GetPersistorForObject(pricerule).GetCollection()
	defer session.Close()

	_, err := collection.Upsert(bson.M{"id": pricerule.ID}, pricerule)

	if err != nil {
		return err
	}
	return nil
}

func (pricerule *PriceRule) checkIfBonusVoucher() {
	//check and fix TypeBonusVoucher
	if pricerule.Type == TypeBonusVoucher {
		// Overwrite values which should not be set differently for bonus voucher promos
		pricerule.Action = ActionCartByAbsolute
		pricerule.MinOrderAmount = 0
		pricerule.MinOrderAmountApplicableItemsOnly = false
		pricerule.ExcludedCustomerGroupIDS = []string{}
		pricerule.IncludedCustomerGroupIDS = []string{}
		pricerule.IsAmountIndependentOfQty = true
		pricerule.Exclusive = false
		pricerule.ItemSets = [][]string{}
		pricerule.MaxUsesPerCustomer = 1
		pricerule.QtyThreshold = 0
		pricerule.ScaledAmounts = []ScaledAmountLevel{}
		pricerule.WhichXYList = []string{}
		pricerule.CalculateDiscountedOrderAmount = true

	}
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
	if Verbose {
		log.Println("updated rule usage history: " + pricerule.ID)
	}
	return pricerule.Upsert()
}

// Delete - delete PriceRule - ID must be set
func (pricerule *PriceRule) Delete() error {
	session, collection := GetPersistorForObject(new(PriceRule)).GetCollection()
	defer session.Close()

	err := collection.Remove(bson.M{"id": pricerule.ID})
	pricerule = nil
	return err
}

func DeletePriceRules(query bson.M) error {
	session, collection := GetPersistorForObject(new(PriceRule)).GetCollection()
	defer session.Close()

	_, err := collection.RemoveAll(query)
	return err
}

// DeletePriceRule - delete PriceRule
func DeletePriceRule(ID string) error {
	session, collection := GetPersistorForObject(new(PriceRule)).GetCollection()
	defer session.Close()

	err := collection.Remove(bson.M{"id": ID})
	return err
}

// RemoveAllPriceRules -
func RemoveAllPriceRules() error {
	session, collection := GetPersistorForObject(new(PriceRule)).GetCollection()
	defer session.Close()

	_, err := collection.RemoveAll(bson.M{})
	return err
}

// GetValidPriceRulesForCheckoutAttributes - find rule for payment method etc etc
// check ValidFrom, ValidTo
func GetValidPriceRulesForCheckoutAttributes(checkoutAttributes []string, customProvider PriceRuleCustomProvider) ([]PriceRule, error) {

	paymentPriceruleTypes := []Type{TypePromotionCustomer, TypePromotionProduct, TypePromotionOrder, TypePaymentMethodDiscount}
	query := bson.M{"type": bson.M{"$in": paymentPriceruleTypes}, "checkoutattributes": bson.M{"$in": checkoutAttributes}, "validfrom": bson.M{"$lte": time.Now()}, "validto": bson.M{"$gte": time.Now()}}
	return getPromotions(query, customProvider)
}

// GetValidPriceRulesForPromotions - find rule for payment
// check ValidFrom, ValidTo
func GetValidPriceRulesForPromotions(priceRuleTypes []Type, customProvider PriceRuleCustomProvider) ([]PriceRule, error) {
	query := bson.M{"type": bson.M{"$in": priceRuleTypes}, "checkoutattributes": bson.M{"$exists": true, "$size": 0}, "validfrom": bson.M{"$lte": time.Now()}, "validto": bson.M{"$gte": time.Now()}}
	return getPromotions(query, customProvider)
}

func getPromotions(query bson.M, customProvider PriceRuleCustomProvider) ([]PriceRule, error) {
	now := time.Now()
	var result []*PriceRule

	session, collection := GetPersistorForObject(new(PriceRule)).GetCollection()
	defer session.Close()

	err := collection.Find(query).Select(nil).Sort("priority").All(&result)
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
