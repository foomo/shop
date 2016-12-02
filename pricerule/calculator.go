package pricerule

import (
	"log"
	"sort"
	"time"

	"github.com/foomo/shop/order"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

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
	OrderItemID                   string
	AppliedDiscounts              []DiscountApplied
	TotalDiscountAmount           float64 // how much the rules would give
	TotalDiscountAmountApplicable float64 // how much the order value permits

	InitialItemPrice                float64
	CurrentItemPrice                float64
	VoucherCalculationBaseItemPrice float64

	Quantity float64

	CustomerPromotionApplied bool //has it previously been applied
	ProductPromotionApplied  bool //has it previously been applied

	StopApplyingDiscounts bool
}

// OrderDiscounts - applied discounts per positionId
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

// ApplyDiscounts applies all possible discounts on order ... if voucherCodes is "" the voucher is not applied
func ApplyDiscounts(order *order.Order, voucherCodes []string, paymentMethod string, roundTo float64) (OrderDiscounts, *OrderDiscountSummary, error) {
	var ruleVoucherPairs []RuleVoucherPair

	now := time.Now()
	//find the groupIds for order items
	productGroupIDsPerPosition := getProductGroupIDsPerPosition(order)

	//find groups for customer
	groupIDsForCustomer := GetGroupsIDSForItem(order.CustomerData.CustomerId, CustomerGroup)
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
			ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: &paymentRule, Voucher: nil})
		}
	}

	orderDiscounts := NewOrderDiscounts(order)
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
		//apply them
		orderDiscounts = calculateRule(orderDiscounts, priceRulePair, order, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
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
			orderDiscounts = calculateRule(orderDiscounts, priceRulePair, order, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		}
	}
	timeTrack(nowAll, "All rules together")
	for _, orderDiscount := range orderDiscounts {
		summary.TotalDiscount += orderDiscount.TotalDiscountAmount
		summary.TotalDiscountApplicable += orderDiscount.TotalDiscountAmountApplicable
		for _, appliedDiscount := range orderDiscount.AppliedDiscounts {
			summary.AppliedPriceRuleIDs = append(summary.AppliedPriceRuleIDs, appliedDiscount.PriceRuleID)
			if len(appliedDiscount.VoucherCode) > 0 {
				summary.AppliedVoucherCodes = append(summary.AppliedVoucherCodes, appliedDiscount.VoucherCode)
				summary.AppliedVoucherIDs = append(summary.AppliedVoucherIDs, appliedDiscount.VoucherID)
				voucherDiscounts, ok := summary.VoucherDiscounts[appliedDiscount.VoucherCode]
				if !ok {
					summary.VoucherDiscounts[appliedDiscount.VoucherCode] = VoucherDiscount{
						Code:           appliedDiscount.VoucherCode,
						ID:             appliedDiscount.VoucherCode,
						DiscountAmount: appliedDiscount.DiscountAmountApplicable,
					}
				} else {
					voucherDiscounts.DiscountAmount += appliedDiscount.DiscountAmountApplicable
				}
			}
		}
	}
	summary.TotalDiscountPercentage = summary.TotalDiscount / getOrderTotal(order) * 100.0
	summary.TotalDiscountApplicablePercentage = summary.TotalDiscountApplicable / getOrderTotal(order) * 100.0

	summary.AppliedPriceRuleIDs = RemoveDuplicates(summary.AppliedPriceRuleIDs)
	summary.AppliedVoucherIDs = RemoveDuplicates(summary.AppliedVoucherIDs)
	summary.AppliedVoucherCodes = RemoveDuplicates(summary.AppliedVoucherCodes)

	return orderDiscounts, summary, nil
}

func calculateRule(orderDiscounts OrderDiscounts, priceRulePair RuleVoucherPair, order *order.Order, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	ok, priceRuleFailReason := validatePriceRuleForOrder(*priceRulePair.Rule, order, productGroupIDsPerPosition, groupIDsForCustomer)
	nowOne := time.Now()

	if ok {
		switch priceRulePair.Rule.Action {
		case ActionItemByAbsolute:
			orderDiscounts = calculateDiscountsItemByAbsolute(order, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		case ActionItemByPercent:
			orderDiscounts = calculateDiscountsItemByPercent(order, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		case ActionCartByAbsolute:
			orderDiscounts = calculateDiscountsCartByAbsolute(order, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		case ActionCartByPercent:
			orderDiscounts = calculateDiscountsCartByPercentage(order, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		case ActionBuyXGetY:
			orderDiscounts = calculateDiscountsBuyXGetY(order, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		case ActionScaled:
			orderDiscounts = calculateScaledDiscounts(order, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		}
		log.Println(":-) Applied " + priceRulePair.Rule.ID)
		timeTrack(nowOne, priceRulePair.Rule.ID)
	} else {
		log.Println(":-/ Not applied " + priceRulePair.Rule.ID + " ----> " + string(priceRuleFailReason))
	}
	return orderDiscounts
}

// find what is the order value of positions that belong to group
// previouslyAppliedDiscounts is for qty 1
func getOrderTotalForPriceRule(priceRule *PriceRule, order *order.Order, productGroupsIDsPerPosition map[string][]string, customerGroupIDs []string) float64 {
	var total float64

	for _, position := range order.Positions {
		productGroupIDs := productGroupsIDsPerPosition[position.ItemID]

		// rule has no customer or product group limitations
		if len(priceRule.IncludedProductGroupIDS) == 0 &&
			len(priceRule.ExcludedProductGroupIDS) == 0 &&
			len(priceRule.IncludedCustomerGroupIDS) == 0 &&
			len(priceRule.ExcludedCustomerGroupIDS) == 0 {
			total += position.Price * position.Quantity
		} else {
			//only sum up if limitations are matched
			if areAllRuleGroupsIncludedInPositionGroups(priceRule.IncludedProductGroupIDS, productGroupIDs) &&
				areAllRuleGroupsNotPresentInPositionGroups(priceRule.ExcludedProductGroupIDS, productGroupIDs) &&
				areAllRuleGroupsIncludedInPositionGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) &&
				areAllRuleGroupsNotPresentInPositionGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
				total += position.Price * position.Quantity
			}
		}
	}
	return total
}

// find what is the order value of positions that belong to group
func getOrderTotal(order *order.Order) float64 {
	var total float64
	for _, position := range order.Positions {
		total += position.Price * position.Quantity
	}
	//fmt.Println("Order total  " + strconv.FormatFloat(total, 'f', 6, 64))
	return total
}

// check if all rule group IDs are included
func areAllRuleGroupsIncludedInPositionGroups(ruleIncludeGroups []string, productAndCustomerGroups []string) bool {
	for _, ruleGroupID := range ruleIncludeGroups {
		if isValueInList(ruleGroupID, productAndCustomerGroups) == false {
			return false
		}
	}
	return true
}

// check if all rule group IDs are included
func areAllRuleGroupsNotPresentInPositionGroups(ruleExcludeGroups []string, productAndCustomerGroups []string) bool {
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

// NewOrderDiscounts - create a new empty map with entries for each position
func NewOrderDiscounts(order *order.Order) OrderDiscounts {
	orderDiscounts := make(OrderDiscounts)
	for _, position := range order.Positions {
		itemDiscountCalculationData := *new(DiscountCalculationData)
		itemDiscountCalculationData.OrderItemID = position.ItemID
		itemDiscountCalculationData.CurrentItemPrice = position.Price
		itemDiscountCalculationData.InitialItemPrice = position.Price
		itemDiscountCalculationData.Quantity = position.Quantity
		itemDiscountCalculationData.AppliedDiscounts = []DiscountApplied{}
		itemDiscountCalculationData.TotalDiscountAmount = 0.0
		itemDiscountCalculationData.TotalDiscountAmountApplicable = 0.0
		itemDiscountCalculationData.StopApplyingDiscounts = false
		itemDiscountCalculationData.VoucherCalculationBaseItemPrice = position.Price
		orderDiscounts[position.ItemID] = itemDiscountCalculationData
	}
	return orderDiscounts
}

// ValidatePriceRuleForOrder -
func validatePriceRuleForOrder(priceRule PriceRule, order *order.Order, productGroupIDsPerPosition map[string][]string, customerGroupIDs []string) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, order, nil, productGroupIDsPerPosition, customerGroupIDs)
}

// ValidatePriceRuleForPosition -
func validatePriceRuleForPosition(priceRule PriceRule, order *order.Order, position *order.Position, productGroupIDsPerPosition map[string][]string, customerGroupIDs []string) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, order, position, productGroupIDsPerPosition, customerGroupIDs)
}

// validatePriceRule -
func validatePriceRule(priceRule PriceRule, order *order.Order, checkedPosition *order.Position, productGroupIDsPerPosition map[string][]string, customerGroupIDs []string) (ok bool, reason TypeRuleValidationMsg) {
	if priceRule.MaxUses <= priceRule.UsageHistory.TotalUsages {
		return false, ValidationPriceRuleMaxUsages
	}
	// if we have the use and the customer history usage ... check ...
	if customerUsages, ok := priceRule.UsageHistory.UsagesPerCustomer[order.CustomerData.CustomerId]; ok && len(order.CustomerData.CustomerId) > 0 {
		if priceRule.MaxUsesPerCustomer <= customerUsages {
			return false, ValidationPriceRuleMaxUsagesPerCustomer
		}
	}

	if priceRule.MinOrderAmount > 0.0 {
		if priceRule.MinOrderAmountApplicableItemsOnly {
			if priceRule.MinOrderAmount > getOrderTotalForPriceRule(&priceRule, order, productGroupIDsPerPosition, customerGroupIDs) {
				return false, ValidationPriceRuleMinimumAmount
			}
		} else {
			if priceRule.MinOrderAmount > getOrderTotal(order) {
				return false, ValidationPriceRuleMinimumAmount
			}
		}
	}

	var productGroupIncludeMatchOK = false
	var productGroupExcludeMatchOK = false
	var customerGroupIncludeMatchOK = false
	var customerGroupExcludeMatchOK = false

	for _, position := range order.Positions {
		if checkedPosition != nil {
			if checkedPosition.ItemID != position.ItemID {
				continue
			}
		}

		if areAllRuleGroupsIncludedInPositionGroups(priceRule.IncludedProductGroupIDS, productGroupIDsPerPosition[position.ItemID]) {
			productGroupIncludeMatchOK = true
		}
		if areAllRuleGroupsNotPresentInPositionGroups(priceRule.ExcludedProductGroupIDS, productGroupIDsPerPosition[position.ItemID]) {
			productGroupExcludeMatchOK = true
		}
		if areAllRuleGroupsIncludedInPositionGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) {
			customerGroupIncludeMatchOK = true
		}

		if areAllRuleGroupsNotPresentInPositionGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
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

// get map of [ItemID] -> [groupID1, groupID2]
func getProductGroupIDsPerPosition(order *order.Order) map[string][]string {
	//product groups per position
	productGroupsPerPosition := make(map[string][]string) //ItemID -> []GroupID
	for _, positionVo := range order.Positions {
		productGroupsPerPosition[positionVo.ItemID] = GetGroupsIDSForItem(positionVo.ItemID, ProductGroup)
	}
	return productGroupsPerPosition
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
