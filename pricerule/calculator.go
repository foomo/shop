package pricerule

import (
	"log"
	"sort"
	"time"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

type ItemCollection struct {
	Items        []*Item
	CustomerType string
	CustomerID   string
}

type Item struct {
	ID         string
	Price      float64
	CrossPrice float64
	Quantity   float64
}

// DiscountApplied -
type DiscountApplied struct {
	PriceRuleID              string
	MappingID                string
	VoucherID                string
	VoucherCode              string
	DiscountAmount           float64
	DiscountSingle           float64
	DiscountAmountApplicable float64
	DiscountSingleApplicable float64

	Quantity             float64
	Price                float64 //price without reductions
	CalculationBasePrice float64 //price used for the calculation of the discount
}

// DiscountCalculationData - of an item
type DiscountCalculationData struct {
	OrderID                       string
	AppliedDiscounts              []DiscountApplied
	TotalDiscountAmount           float64 // how much the rules would give
	TotalDiscountAmountApplicable float64 // how much the itemCollection value permits

	InitialItemPrice                float64
	CurrentItemPrice                float64
	VoucherCalculationBaseItemPrice float64

	Quantity float64

	CustomerPromotionApplied bool //has it previously been applied
	ProductPromotionApplied  bool //has it previously been applied

	StopApplyingDiscounts bool
}

// OrderDiscounts - applied discounts per itemId
type OrderDiscounts map[string]DiscountCalculationData

type VoucherDiscount struct {
	Code           string
	ID             string
	DiscountAmount float64
}

// OrderDiscountSummary -
type OrderDiscountSummary struct {
	AppliedPriceRuleIDs               []string
	AppliedVoucherIDs                 []string
	AppliedVoucherCodes               []string
	TotalDiscount                     float64
	TotalDiscountPercentage           float64
	TotalDiscountApplicable           float64
	TotalDiscountApplicablePercentage float64
	VoucherDiscounts                  map[string]VoucherDiscount
}

// RuleVoucherPair -
type RuleVoucherPair struct {
	Rule    *PriceRule
	Voucher *Voucher
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// ApplyDiscounts applies all possible discounts on itemCollection ... if voucherCodes is "" the voucher is not applied
func ApplyDiscounts(itemCollection *ItemCollection, voucherCodes []string, paymentMethod string, roundTo float64) (OrderDiscounts, *OrderDiscountSummary, error) {
	var ruleVoucherPairs []RuleVoucherPair

	now := time.Now()
	//find the groupIds for itemCollection items
	productGroupIDsPerItem := getProductGroupIDsPerItem(itemCollection)

	//find groups for customer
	groupIDsForCustomer := GetGroupsIDSForItem(itemCollection.CustomerType, CustomerGroup)
	if len(groupIDsForCustomer) == 0 {
		groupIDsForCustomer = []string{}
	}

	// find applicable pricerules - auto promotions
	promotionPriceRules, err := GetValidPriceRulesForPromotions([]Type{TypePromotionCustomer, TypePromotionProduct, TypePromotionOrder})

	if err != nil {
		return nil, nil, err
	}

	for _, promotionRule := range promotionPriceRules {
		rule := &PriceRule{}
		*rule = promotionRule
		ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: rule, Voucher: nil})
	}

	// find applicable payment discounts
	var paymentPriceRules []PriceRule
	if len(paymentMethod) > 0 {
		paymentPriceRules, err = GetValidPriceRulesForPaymentMethod(paymentMethod)
		if err != nil {
			return nil, nil, err
		}
		for _, paymentRule := range paymentPriceRules {

			rule := &PriceRule{}
			*rule = paymentRule
			ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: rule, Voucher: nil})
		}
	}

	itemCollDiscounts := NewOrderDiscounts(itemCollection)
	summary := &OrderDiscountSummary{
		AppliedVoucherCodes: []string{},
		AppliedVoucherIDs:   []string{},
		VoucherDiscounts:    map[string]VoucherDiscount{},
	}
	timeTrack(now, "peparations took ")
	nowAll := time.Now()

	// ~ PRICERULE PAIR SORTING - BY PRIORITY - higher priority means it is applied first
	sort.Sort(ByPriority(ruleVoucherPairs))

	//first loop where all promotion discounts are applied

	for _, priceRulePair := range ruleVoucherPairs {
		pair := RuleVoucherPair{}
		pair = priceRulePair
		//apply them
		itemCollDiscounts = calculateRule(itemCollDiscounts, pair, itemCollection, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
	}

	//find the vouchers and voucher rules
	//find applicable pricerules of type TypeVoucher for

	if len(voucherCodes) > 0 {
		var ruleVoucherPairsStep2 []RuleVoucherPair
		for _, voucherCode := range voucherCodes {
			if len(voucherCode) > 0 {
				voucherVo, voucherPriceRule, err := GetVoucherAndPriceRule(voucherCode)
				if voucherVo == nil {
					log.Println("voucher not found for code: " + voucherCode + " in " + "priceRule.ApplyDiscounts")
				}
				if err != nil {
					log.Println(err)
					log.Println("skupping voucher " + voucherCode)
					continue
				}

				if !voucherVo.TimeRedeemed.IsZero() {
					log.Println("voucher " + voucherCode + " already redeemed ... skipping")
					continue
				}

				pair := RuleVoucherPair{
					Rule:    voucherPriceRule,
					Voucher: voucherVo,
				}
				ruleVoucherPairsStep2 = append(ruleVoucherPairsStep2, pair)
			}
		}
		//apply them

		for _, priceRulePair := range ruleVoucherPairsStep2 {
			itemCollDiscounts = calculateRule(itemCollDiscounts, priceRulePair, itemCollection, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		}
	}
	timeTrack(nowAll, "All rules together")
	for _, itemCollDiscount := range itemCollDiscounts {
		summary.TotalDiscount += itemCollDiscount.TotalDiscountAmount
		summary.TotalDiscountApplicable += itemCollDiscount.TotalDiscountAmountApplicable
		for _, appliedDiscount := range itemCollDiscount.AppliedDiscounts {
			summary.AppliedPriceRuleIDs = append(summary.AppliedPriceRuleIDs, appliedDiscount.PriceRuleID)
			if len(appliedDiscount.VoucherCode) > 0 {
				summary.AppliedVoucherCodes = append(summary.AppliedVoucherCodes, appliedDiscount.VoucherCode)
				summary.AppliedVoucherIDs = append(summary.AppliedVoucherIDs, appliedDiscount.VoucherID)
				log.Println("#### voucherCode: ", appliedDiscount.VoucherCode)
				log.Println("#### voucherID: ", appliedDiscount.VoucherID)

				voucherDiscounts, ok := summary.VoucherDiscounts[appliedDiscount.VoucherCode]
				if !ok {
					summary.VoucherDiscounts[appliedDiscount.VoucherCode] = VoucherDiscount{
						Code:           appliedDiscount.VoucherCode,
						ID:             appliedDiscount.VoucherCode,
						DiscountAmount: appliedDiscount.DiscountAmountApplicable,
					}
				} else {
					voucherDiscounts.DiscountAmount += appliedDiscount.DiscountAmountApplicable
					summary.VoucherDiscounts[appliedDiscount.VoucherCode] = voucherDiscounts
				}

			}
		}
	}
	summary.TotalDiscountPercentage = summary.TotalDiscount / getOrderTotal(itemCollection) * 100.0
	summary.TotalDiscountApplicablePercentage = summary.TotalDiscountApplicable / getOrderTotal(itemCollection) * 100.0

	summary.AppliedPriceRuleIDs = RemoveDuplicates(summary.AppliedPriceRuleIDs)
	summary.AppliedVoucherIDs = RemoveDuplicates(summary.AppliedVoucherIDs)
	summary.AppliedVoucherCodes = RemoveDuplicates(summary.AppliedVoucherCodes)

	return itemCollDiscounts, summary, nil
}

func calculateRule(itemCollDiscounts OrderDiscounts, priceRulePair RuleVoucherPair, itemCollection *ItemCollection, productGroupIDsPerItem map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	ok, priceRuleFailReason := validatePriceRuleForOrder(*priceRulePair.Rule, itemCollection, productGroupIDsPerItem, groupIDsForCustomer)
	nowOne := time.Now()

	if ok {
		switch priceRulePair.Rule.Action {
		case ActionItemByAbsolute:
			itemCollDiscounts = calculateDiscountsItemByAbsolute(itemCollection, priceRulePair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		case ActionItemByPercent:
			itemCollDiscounts = calculateDiscountsItemByPercent(itemCollection, priceRulePair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		case ActionCartByAbsolute:
			itemCollDiscounts = calculateDiscountsCartByAbsolute(itemCollection, priceRulePair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		case ActionCartByPercent:
			itemCollDiscounts = calculateDiscountsCartByPercentage(itemCollection, priceRulePair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		case ActionBuyXGetY:
			itemCollDiscounts = calculateDiscountsBuyXGetY(itemCollection, priceRulePair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		case ActionScaled:
			itemCollDiscounts = calculateScaledDiscounts(itemCollection, priceRulePair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		}
		log.Println(":-) Applied " + priceRulePair.Rule.ID)
		timeTrack(nowOne, priceRulePair.Rule.ID)
	} else {
		log.Println(":-/ Not applied " + priceRulePair.Rule.ID + " ----> " + string(priceRuleFailReason))
	}
	return itemCollDiscounts
}

// find what is the itemCollection value of items that belong to group
// previouslyAppliedDiscounts is for qty 1
func getOrderTotalForPriceRule(priceRule *PriceRule, itemCollection *ItemCollection, productGroupsIDsPerItem map[string][]string, customerGroupIDs []string) float64 {
	var total float64

	for _, item := range itemCollection.Items {
		productGroupIDs := productGroupsIDsPerItem[item.ID]

		// rule has no customer or product group limitations
		if len(priceRule.IncludedProductGroupIDS) == 0 &&
			len(priceRule.ExcludedProductGroupIDS) == 0 &&
			len(priceRule.IncludedCustomerGroupIDS) == 0 &&
			len(priceRule.ExcludedCustomerGroupIDS) == 0 {
			total += item.Price * item.Quantity
		} else {
			//only sum up if limitations are matched
			if areAllRuleGroupsIncludedInItemGroups(priceRule.IncludedProductGroupIDS, productGroupIDs) &&
				areAllRuleGroupsNotPresentInItemGroups(priceRule.ExcludedProductGroupIDS, productGroupIDs) &&
				areAllRuleGroupsIncludedInItemGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) &&
				areAllRuleGroupsNotPresentInItemGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
				total += item.Price * item.Quantity
			}
		}
	}
	return total
}

// find what is the itemCollection value of items that belong to group
func getOrderTotal(itemCollection *ItemCollection) float64 {
	var total float64
	for _, item := range itemCollection.Items {
		total += item.Price * item.Quantity
	}
	//fmt.Println("Order total  " + strconv.FormatFloat(total, 'f', 6, 64))
	return total
}

// check if all rule group IDs are included
func areAllRuleGroupsIncludedInItemGroups(ruleIncludeGroups []string, productAndCustomerGroups []string) bool {
	for _, ruleGroupID := range ruleIncludeGroups {
		if isValueInList(ruleGroupID, productAndCustomerGroups) == false {
			return false
		}
	}
	return true
}

// check if all rule group IDs are included
func areAllRuleGroupsNotPresentInItemGroups(ruleExcludeGroups []string, productAndCustomerGroups []string) bool {
	for _, ruleGroupID := range ruleExcludeGroups {
		if isValueInList(ruleGroupID, productAndCustomerGroups) == true {
			return false
		}
	}
	return true
}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// NewOrderDiscounts - create a new empty map with entries for each item
func NewOrderDiscounts(itemCollection *ItemCollection) OrderDiscounts {
	itemCollDiscounts := make(OrderDiscounts)
	for _, item := range itemCollection.Items {
		itemDiscountCalculationData := *new(DiscountCalculationData)
		itemDiscountCalculationData.OrderID = item.ID
		itemDiscountCalculationData.CurrentItemPrice = item.Price
		itemDiscountCalculationData.InitialItemPrice = item.Price
		itemDiscountCalculationData.Quantity = item.Quantity
		itemDiscountCalculationData.AppliedDiscounts = []DiscountApplied{}
		itemDiscountCalculationData.TotalDiscountAmount = 0.0
		itemDiscountCalculationData.TotalDiscountAmountApplicable = 0.0
		itemDiscountCalculationData.StopApplyingDiscounts = false
		itemDiscountCalculationData.VoucherCalculationBaseItemPrice = item.Price
		itemCollDiscounts[item.ID] = itemDiscountCalculationData
	}
	return itemCollDiscounts
}

// ValidatePriceRuleForOrder -
func validatePriceRuleForOrder(priceRule PriceRule, itemCollection *ItemCollection, productGroupIDsPerItem map[string][]string, customerGroupIDs []string) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, itemCollection, nil, productGroupIDsPerItem, customerGroupIDs)
}

// ValidatePriceRuleForItem -
func validatePriceRuleForItem(priceRule PriceRule, itemCollection *ItemCollection, item *Item, productGroupIDsPerItem map[string][]string, customerGroupIDs []string) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, itemCollection, item, productGroupIDsPerItem, customerGroupIDs)
}

// validatePriceRule -
func validatePriceRule(priceRule PriceRule, itemCollection *ItemCollection, checkedItem *Item, productGroupIDsPerItem map[string][]string, customerGroupIDs []string) (ok bool, reason TypeRuleValidationMsg) {
	if priceRule.MaxUses <= priceRule.UsageHistory.TotalUsages {
		return false, ValidationPriceRuleMaxUsages
	}
	// if we have the use and the customer history usage ... check ...
	if customerUsages, ok := priceRule.UsageHistory.UsagesPerCustomer[itemCollection.CustomerID]; ok && len(itemCollection.CustomerID) > 0 {
		if priceRule.MaxUsesPerCustomer <= customerUsages {
			return false, ValidationPriceRuleMaxUsagesPerCustomer
		}
	}

	if priceRule.MinOrderAmount > 0.0 {
		if priceRule.MinOrderAmountApplicableItemsOnly {
			if priceRule.MinOrderAmount > getOrderTotalForPriceRule(&priceRule, itemCollection, productGroupIDsPerItem, customerGroupIDs) {
				return false, ValidationPriceRuleMinimumAmount
			}
		} else {
			if priceRule.MinOrderAmount > getOrderTotal(itemCollection) {
				return false, ValidationPriceRuleMinimumAmount
			}
		}
	}

	var productGroupIncludeMatchOK = false
	var productGroupExcludeMatchOK = false
	var customerGroupIncludeMatchOK = false
	var customerGroupExcludeMatchOK = false

	for _, item := range itemCollection.Items {
		if checkedItem != nil {
			if checkedItem.ID != item.ID {
				continue
			}
		}

		if areAllRuleGroupsIncludedInItemGroups(priceRule.IncludedProductGroupIDS, productGroupIDsPerItem[item.ID]) {
			productGroupIncludeMatchOK = true
		}
		if areAllRuleGroupsNotPresentInItemGroups(priceRule.ExcludedProductGroupIDS, productGroupIDsPerItem[item.ID]) {
			productGroupExcludeMatchOK = true
		}
		if areAllRuleGroupsIncludedInItemGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) {
			customerGroupIncludeMatchOK = true
		}

		if areAllRuleGroupsNotPresentInItemGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
			customerGroupExcludeMatchOK = true
		}
	}

	if !productGroupIncludeMatchOK {
		return false, ValidationPriceRuleIncludeProductGroupsNotMatching
	}

	if !productGroupExcludeMatchOK {
		return false, ValidationPriceRuleExcludeProductGroupsNotMatching
	}

	if !customerGroupIncludeMatchOK {
		return false, ValidationPriceRuleIncludeCustomerGroupsNotMatching
	}

	if !customerGroupExcludeMatchOK {
		return false, ValidationPriceRuleExcludeCustomerGroupsNotMatching
	}
	return true, ValidationPriceRuleOK
}

// get map of [ID] -> [groupID1, groupID2]
func getProductGroupIDsPerItem(itemCollection *ItemCollection) map[string][]string {
	//product groups per item
	productGroupsPerItem := make(map[string][]string) //ID -> []GroupID
	for _, itemVo := range itemCollection.Items {
		productGroupsPerItem[itemVo.ID] = GetGroupsIDSForItem(itemVo.ID, ProductGroup)
	}
	return productGroupsPerItem
}

// RoundTo -
func roundToStep(x, unit float64) float64 {
	if x > 0 {
		return float64(int64(x/unit+0.5)) * unit
	}
	return float64(int64(x/unit-0.5)) * unit
}

func dereferenceVoucherPriceRule(voucherRule *PriceRule) PriceRule {
	if voucherRule != nil {
		return *voucherRule
	}
	return *new(PriceRule)
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
