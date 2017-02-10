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

type ArticleCollection struct {
	Articles     []*Article
	CustomerType string
	CustomerID   string
}

type Article struct {
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

	Custom interface{}
}

// DiscountCalculationData - of an item
type DiscountCalculationData struct {
	OrderItemID                   string
	AppliedDiscounts              []DiscountApplied
	TotalDiscountAmount           float64 // how much the rules would give
	TotalDiscountAmountApplicable float64 // how much the articleCollection value permits

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

var cache = NewCache()

func InitCache() {
	cache.InitCatalogCalculationCache()
}

func ClearCache() {
	cache.ClearCatalogCalculationCache()
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// ApplyDiscounts applies all possible discounts on articleCollection ... if voucherCodes is "" the voucher is not applied
// This is not yet used. ApplyDiscounts should at some point be able to consider previousle calculated discounts
func ApplyDiscounts(articleCollection *ArticleCollection, existingDiscounts OrderDiscounts, voucherCodes []string, paymentMethod string, roundTo float64, customProvider PriceRuleCustomProvider) (OrderDiscounts, *OrderDiscountSummary, error) {
	var ruleVoucherPairs []RuleVoucherPair
	now := time.Now()
	//find the groupIds for articleCollection items
	productGroupIDsPerPosition := getProductGroupIDsPerPosition(articleCollection, false)

	//find groups for customer
	groupIDsForCustomer := GetGroupsIDSForItem(articleCollection.CustomerType, CustomerGroup)
	if len(groupIDsForCustomer) == 0 {
		groupIDsForCustomer = []string{}
	}

	timeTrack(now, "groups data took ")
	// find applicable pricerules - auto promotions
	promotionPriceRules, err := GetValidPriceRulesForPromotions([]Type{TypePromotionCustomer, TypePromotionProduct, TypePromotionOrder}, customProvider)

	if err != nil {
		return nil, nil, err
	}

	for _, promotionRule := range promotionPriceRules {
		rule := &PriceRule{}
		*rule = promotionRule
		ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: rule, Voucher: nil})
	}

	timeTrack(now, "loading pricerules took ")

	// find applicable payment discounts
	var paymentPriceRules []PriceRule
	if len(paymentMethod) > 0 {
		paymentPriceRules, err = GetValidPriceRulesForPaymentMethod(paymentMethod, customProvider)
		if err != nil {
			return nil, nil, err
		}
		for _, paymentRule := range paymentPriceRules {

			rule := &PriceRule{}
			*rule = paymentRule
			ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: rule, Voucher: nil})
		}
	}
	timeTrack(now, "loading pricerules took ")

	orderDiscounts := NewOrderDiscounts(articleCollection)
	summary := &OrderDiscountSummary{
		AppliedVoucherCodes: []string{},
		AppliedVoucherIDs:   []string{},
		VoucherDiscounts:    map[string]VoucherDiscount{},
	}
	timeTrack(now, "preparations took ")
	nowAll := time.Now()

	// ~ PRICERULE PAIR SORTING - BY PRIORITY - higher priority means it is applied first
	sort.Sort(ByPriority(ruleVoucherPairs))

	//first loop where all promotion discounts are applied

	for _, priceRulePair := range ruleVoucherPairs {
		pair := RuleVoucherPair{}
		pair = priceRulePair
		//apply them
		orderDiscounts = calculateRule(orderDiscounts, pair, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, false)
	}

	//find the vouchers and voucher rules
	//find applicable pricerules of type TypeVoucher for

	if len(voucherCodes) > 0 {
		var ruleVoucherPairsStep2 []RuleVoucherPair
		for _, voucherCode := range voucherCodes {
			if len(voucherCode) > 0 {
				voucherVo, voucherPriceRule, err := GetVoucherAndPriceRule(voucherCode, customProvider)
				if voucherVo == nil {
					log.Println("voucher not found for code: " + voucherCode + " in " + "priceRule.ApplyDiscounts")
				}
				if err != nil {
					log.Println("skipping voucher "+voucherCode, err)
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
			orderDiscounts = calculateRule(orderDiscounts, priceRulePair, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, false)
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
	summary.TotalDiscountPercentage = summary.TotalDiscount / getOrderTotal(articleCollection) * 100.0
	summary.TotalDiscountApplicablePercentage = summary.TotalDiscountApplicable / getOrderTotal(articleCollection) * 100.0
	summary.AppliedPriceRuleIDs = RemoveDuplicates(summary.AppliedPriceRuleIDs)
	summary.AppliedVoucherIDs = RemoveDuplicates(summary.AppliedVoucherIDs)
	summary.AppliedVoucherCodes = RemoveDuplicates(summary.AppliedVoucherCodes)
	return orderDiscounts, summary, nil
}

// ApplyDiscountsOnCatalog applies all possible discounts on articleCollection ... if voucherCodes is "" the voucher is not applied
// This is not yet used. ApplyDiscounts should at some point be able to consider previousle calculated discounts
func ApplyDiscountsOnCatalog(articleCollection *ArticleCollection, existingDiscounts OrderDiscounts, voucherCodes []string, paymentMethod string, roundTo float64, customProvider PriceRuleCustomProvider) (OrderDiscounts, *OrderDiscountSummary, error) {
	var ruleVoucherPairs []RuleVoucherPair
	now := time.Now()
	start := time.Now()

	now = time.Now()
	//find the groupIds for articleCollection items
	productGroupIDsPerPosition := getProductGroupIDsPerPosition(articleCollection, true)
	timeTrack(now, "[ApplyDiscountsOnCatalog] loading of productGroupIDsPerPosition took ")
	now = time.Now()
	//find groups for customer
	groupIDsForCustomer := cache.GetGroupsIDSForItem(articleCollection.CustomerType, CustomerGroup)
	if len(groupIDsForCustomer) == 0 {
		groupIDsForCustomer = []string{}
	}
	timeTrack(now, "[ApplyDiscountsOnCatalog] loading of groupIDsForCustomer took ")
	now = time.Now()

	// find applicable pricerules - auto promotions
	promotionPriceRules, err := cache.CachedGetValidProductAndCustomerPriceRules(customProvider)

	timeTrack(now, "[ApplyDiscountsOnCatalog] loading pricerules took ")
	now = time.Now()

	if err != nil {
		return nil, nil, err
	}

	for _, promotionRule := range promotionPriceRules {
		rule := &PriceRule{}
		*rule = promotionRule
		ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: rule, Voucher: nil})
	}

	orderDiscounts := NewOrderDiscounts(articleCollection)
	summary := &OrderDiscountSummary{
		AppliedVoucherCodes: []string{},
		AppliedVoucherIDs:   []string{},
		VoucherDiscounts:    map[string]VoucherDiscount{},
	}
	timeTrack(now, "[ApplyDiscountsOnCatalog] preparations LAST STEP took ")
	timeTrack(start, "[ApplyDiscountsOnCatalog] preparations took ")

	// ~ PRICERULE PAIR SORTING - BY PRIORITY - higher priority means it is applied first
	sort.Sort(ByPriority(ruleVoucherPairs))

	//first loop where all promotion discounts are applied
	now = time.Now()
	for _, priceRulePair := range ruleVoucherPairs {
		pair := RuleVoucherPair{}
		pair = priceRulePair
		//apply them
		orderDiscounts = calculateRule(orderDiscounts, pair, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, true)
	}

	timeTrack(now, "[ApplyDiscountsOnCatalog] CALCULATIONS and CHECKS took ")
	for _, orderDiscount := range orderDiscounts {
		//summary.TotalDiscount += orderDiscount.TotalDiscountAmount
		//summary.TotalDiscountApplicable += orderDiscount.TotalDiscountAmountApplicable
		for _, appliedDiscount := range orderDiscount.AppliedDiscounts {
			summary.AppliedPriceRuleIDs = append(summary.AppliedPriceRuleIDs, appliedDiscount.PriceRuleID)
		}
	}
	//summary.TotalDiscountPercentage = summary.TotalDiscount / getOrderTotal(articleCollection) * 100.0
	//summary.TotalDiscountApplicablePercentage = summary.TotalDiscountApplicable / getOrderTotal(articleCollection) * 100.0

	summary.AppliedPriceRuleIDs = RemoveDuplicates(summary.AppliedPriceRuleIDs)
	summary.AppliedVoucherIDs = []string{}
	summary.AppliedVoucherCodes = []string{}
	timeTrack(start, "[ApplyDiscountsOnCatalog] ALL DONE took ")
	return orderDiscounts, summary, nil
}

func calculateRule(orderDiscounts OrderDiscounts, priceRulePair RuleVoucherPair, articleCollection *ArticleCollection, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64, isCatalogCalculation bool) OrderDiscounts {
	ok := true
	if isCatalogCalculation == false {
		ok, _ = validatePriceRuleForOrder(*priceRulePair.Rule, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, isCatalogCalculation)
	} else {
		//bypass order check if catalog computation
		ok = true
	}

	if ok == true {
		switch priceRulePair.Rule.Action {
		case ActionItemByAbsolute:
			orderDiscounts = calculateDiscountsItemByAbsolute(articleCollection, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		case ActionItemByPercent:
			orderDiscounts = calculateDiscountsItemByPercent(articleCollection, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		case ActionCartByAbsolute:
			orderDiscounts = calculateDiscountsCartByAbsolute(articleCollection, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		case ActionCartByPercent:
			orderDiscounts = calculateDiscountsCartByPercentage(articleCollection, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		case ActionBuyXPayY:
			orderDiscounts = calculateDiscountsBuyXPayY(articleCollection, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		case ActionScaled:
			orderDiscounts = calculateScaledDiscounts(articleCollection, priceRulePair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		}
	}
	return orderDiscounts
}

// find what is the articleCollection value of positions that belong to group
// previouslyAppliedDiscounts is for qty 1
func getOrderTotalForPriceRule(priceRule *PriceRule, articleCollection *ArticleCollection, productGroupsIDsPerPosition map[string][]string, customerGroupIDs []string, orderDiscounts OrderDiscounts) float64 {
	var total float64

	for _, article := range articleCollection.Articles {
		productGroupIDs := productGroupsIDsPerPosition[article.ID]

		// rule has no customer or product group limitations
		if len(priceRule.IncludedProductGroupIDS) == 0 &&
			len(priceRule.ExcludedProductGroupIDS) == 0 &&
			len(priceRule.IncludedCustomerGroupIDS) == 0 &&
			len(priceRule.ExcludedCustomerGroupIDS) == 0 {
			total += article.Price * article.Quantity
		} else {
			//only sum up if limitations are matched
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, productGroupIDs) &&
				IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, productGroupIDs) &&
				IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) &&
				IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
				sub := 0.0
				if orderDiscounts != nil {
					previouslyAppliedDiscounts, ok := orderDiscounts[article.ID]
					if ok {
						sub = previouslyAppliedDiscounts.TotalDiscountAmount
					}
				}
				total += article.Price*article.Quantity - sub
			}
		}
	}
	return total
}

// find what is the articleCollection value of positions that belong to group
func getOrderTotal(articleCollection *ArticleCollection) float64 {
	var total float64
	for _, article := range articleCollection.Articles {
		total += article.Price * article.Quantity
	}
	return total
}

// check if all rule group IDs are included
func IsOneProductOrCustomerGroupInIncludedGroups(ruleIncludeGroups []string, productAndCustomerGroups []string) bool {
	if len(ruleIncludeGroups) == 0 {
		return true
	}
	for _, p := range productAndCustomerGroups {
		if isValueInList(p, ruleIncludeGroups) {

			return true
		}
	}
	return false

}

// check if all rule group IDs are included
func IsNoProductOrGroupInExcludeGroups(ruleExcludeGroups []string, productAndCustomerGroups []string) bool {

	if len(ruleExcludeGroups) == 0 {
		return true
	}
	for _, p := range productAndCustomerGroups {
		if isValueInList(p, ruleExcludeGroups) {
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

// NewOrderDiscounts - create a new empty map with entries for each article
func NewOrderDiscounts(articleCollection *ArticleCollection) OrderDiscounts {
	orderDiscounts := make(OrderDiscounts)
	for _, article := range articleCollection.Articles {
		itemDiscountCalculationData := *new(DiscountCalculationData)
		itemDiscountCalculationData.OrderItemID = article.ID
		itemDiscountCalculationData.CurrentItemPrice = article.Price
		itemDiscountCalculationData.InitialItemPrice = article.Price
		itemDiscountCalculationData.Quantity = article.Quantity
		itemDiscountCalculationData.AppliedDiscounts = []DiscountApplied{}
		itemDiscountCalculationData.TotalDiscountAmount = 0.0
		itemDiscountCalculationData.TotalDiscountAmountApplicable = 0.0
		itemDiscountCalculationData.StopApplyingDiscounts = false
		itemDiscountCalculationData.VoucherCalculationBaseItemPrice = article.Price
		orderDiscounts[article.ID] = itemDiscountCalculationData
	}
	return orderDiscounts
}

// ValidatePriceRuleForOrder -
func validatePriceRuleForOrder(priceRule PriceRule, articleCollection *ArticleCollection, productGroupIDsPerPosition map[string][]string, customerGroupIDs []string, isCatalogCalculation bool) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, articleCollection, nil, productGroupIDsPerPosition, customerGroupIDs, isCatalogCalculation)
}

// ValidatePriceRuleForPosition -
func validatePriceRuleForPosition(priceRule PriceRule, articleCollection *ArticleCollection, article *Article, productGroupIDsPerPosition map[string][]string, customerGroupIDs []string, isCatalogCalculation bool) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, articleCollection, article, productGroupIDsPerPosition, customerGroupIDs, isCatalogCalculation)
}

// validatePriceRule -
func validatePriceRule(priceRule PriceRule, articleCollection *ArticleCollection, checkedPosition *Article, productGroupIDsPerPosition map[string][]string, customerGroupIDs []string, isCatalogCalculation bool) (ok bool, reason TypeRuleValidationMsg) {
	if !isCatalogCalculation {
		if priceRule.MaxUses <= priceRule.UsageHistory.TotalUsages {
			return false, ValidationPriceRuleMaxUsages
		}
		// if we have the use and the customer history usage ... check ...
		if customerUsages, ok := priceRule.UsageHistory.UsagesPerCustomer[articleCollection.CustomerID]; ok && len(articleCollection.CustomerID) > 0 {
			if priceRule.MaxUsesPerCustomer <= customerUsages {
				return false, ValidationPriceRuleMaxUsagesPerCustomer
			}
		}

		if priceRule.MinOrderAmount > 0.0 {
			if priceRule.MinOrderAmountApplicableItemsOnly {
				if priceRule.MinOrderAmount > getOrderTotalForPriceRule(&priceRule, articleCollection, productGroupIDsPerPosition, customerGroupIDs, nil) {
					return false, ValidationPriceRuleMinimumAmount
				}
			} else {
				if priceRule.MinOrderAmount > getOrderTotal(articleCollection) {
					return false, ValidationPriceRuleMinimumAmount
				}
			}

		}
	}

	now := time.Now()
	var productGroupIncludeMatchOK = false
	var productGroupExcludeMatchOK = false
	var customerGroupIncludeMatchOK = false
	var customerGroupExcludeMatchOK = false

	if checkedPosition == nil {
		for _, article := range articleCollection.Articles {
			//the continue here is now obsolete
			if checkedPosition != nil {
				if checkedPosition.ID != article.ID {
					continue
				}
			}
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, productGroupIDsPerPosition[article.ID]) {
				productGroupIncludeMatchOK = true
			}
			if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, productGroupIDsPerPosition[article.ID]) {
				productGroupExcludeMatchOK = true
			}
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) {
				customerGroupIncludeMatchOK = true
			}

			if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
				customerGroupExcludeMatchOK = true
			}
		}
	} else {
		// if only checking for one item, do not go through loop
		if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, productGroupIDsPerPosition[checkedPosition.ID]) {
			productGroupIncludeMatchOK = true
		}
		if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, productGroupIDsPerPosition[checkedPosition.ID]) {
			productGroupExcludeMatchOK = true
		}
		if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, customerGroupIDs) {
			customerGroupIncludeMatchOK = true
		}

		if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, customerGroupIDs) {
			customerGroupExcludeMatchOK = true
		}
	}

	if checkedPosition == nil {
		timeTrack(now, "Time spend checking include/exclude groups for all items")
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
func getProductGroupIDsPerPosition(articleCollection *ArticleCollection, isCatalogCalculation bool) map[string][]string {
	//product groups per article

	productGroupsPerPosition := make(map[string][]string) //ItemID -> []GroupID
	for _, positionVo := range articleCollection.Articles {
		if isCatalogCalculation == true {
			productGroupsPerPosition[positionVo.ID] = cache.GetGroupsIDSForItem(positionVo.ID, ProductGroup)
		} else {
			productGroupsPerPosition[positionVo.ID] = GetGroupsIDSForItem(positionVo.ID, ProductGroup)
		}
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
	log.Printf("%s %s", name, elapsed)
}
