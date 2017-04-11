package pricerule

import (
	"log"
	"sort"
	"time"

	"github.com/davecgh/go-spew/spew"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

// InputData

type CalculationParameters struct {
	articleCollection                   *ArticleCollection
	productGroupIDsPerPosition          map[string][]string
	groupIDsForCustomer                 []string
	roundTo                             float64
	isCatalogCalculation                bool
	checkoutAttributes                  []string
	bestOptionCustomeProductRulePerItem map[string]string // which is the product or customer type rule that is applied on item
	blacklistedItemIDs                  []string
}

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

	//helper values for easier rendering
	AppliedInCatalog        bool // applied in catalog calculation
	ApplicableInCatalog     bool //could have been applied in catalog
	IsTypePromotionCustomer bool // is the type TypePromotionCustomer

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
func ApplyDiscounts(articleCollection *ArticleCollection, existingDiscounts OrderDiscounts, voucherCodes []string, checkoutAttributes []string, roundTo float64, customProvider PriceRuleCustomProvider) (OrderDiscounts, *OrderDiscountSummary, error) {
	calculationParameters := &CalculationParameters{}
	calculationParameters.articleCollection = articleCollection
	calculationParameters.roundTo = roundTo
	calculationParameters.isCatalogCalculation = false
	calculationParameters.checkoutAttributes = checkoutAttributes

	now := time.Now()
	//find the groupIds for articleCollection items

	calculationParameters.productGroupIDsPerPosition = getProductGroupIDsPerPosition(articleCollection, false)
	//find groups for customer
	groupIDsForCustomer := GetGroupsIDSForItem(articleCollection.CustomerType, CustomerGroup)
	if len(groupIDsForCustomer) == 0 {
		groupIDsForCustomer = []string{}
	}
	calculationParameters.groupIDsForCustomer = groupIDsForCustomer

	//find blacklisted items
	blacklistedItemIDs, blacklistedItemsErr := GetBlacklistedItemIds()
	if blacklistedItemsErr != nil {
		return nil, nil, blacklistedItemsErr
	}
	calculationParameters.blacklistedItemIDs = blacklistedItemIDs

	// ----------------------------------------------------------------------------------------------------------
	// promotions - step 1
	timeTrack(now, "groups data took ")
	// find applicable pricerules - auto promotions
	promotionPriceRules, err := GetValidPriceRulesForPromotions([]Type{TypePromotionCustomer, TypePromotionProduct, TypePromotionOrder}, customProvider)

	if err != nil {
		return nil, nil, err
	}
	var ruleVoucherPairs []RuleVoucherPair
	for _, promotionRule := range promotionPriceRules {
		rule := &PriceRule{}
		*rule = promotionRule
		ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: rule, Voucher: nil})
	}

	timeTrack(now, "loading pricerules took ")

	// find applicable discounts limited to checkoutAttributes - payment methods etc
	var paymentPriceRules []PriceRule
	if len(checkoutAttributes) > 0 {
		paymentPriceRules, err = GetValidPriceRulesForCheckoutAttributes(checkoutAttributes, customProvider)
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

	bestOptionCustomerProductRulePerItem := getBestOptionCustomerProductRulePerItem(ruleVoucherPairs, calculationParameters)
	calculationParameters.bestOptionCustomeProductRulePerItem = bestOptionCustomerProductRulePerItem
	spew.Dump(bestOptionCustomerProductRulePerItem)

	for _, priceRulePair := range ruleVoucherPairs {
		pair := RuleVoucherPair{}
		pair = priceRulePair
		//apply them
		orderDiscounts = calculateRule(orderDiscounts, pair, calculationParameters)
	}

	// ----------------------------------------------------------------------------------------------------------
	// vouchers: step 2
	// find the vouchers and voucher rules
	// find applicable pricerules of type TypeVoucher for

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

				//filter out the vouchers that can not be applied due to a mismatch with checkoutAttributes
				if len(voucherPriceRule.CheckoutAttributes) > 0 {
					match := false
					for _, checkoutAttribute := range checkoutAttributes {
						if contains(checkoutAttribute, voucherPriceRule.CheckoutAttributes) {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}

				if !voucherVo.TimeRedeemed.IsZero() {
					if Verbose {
						log.Println("voucher " + voucherCode + " already redeemed ... skipping")
					}
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
			orderDiscounts = calculateRule(orderDiscounts, priceRulePair, calculationParameters)
		}
	}

	// ----------------------------------------------------------------------------------------------------------
	// shipping - step 3
	// shipping costs handling
	// find applicable pricerules - auto promotions
	shippingPriceRules, err := GetValidPriceRulesForPromotions([]Type{TypeShipping}, customProvider)

	if err != nil {
		return nil, nil, err
	}

	var ruleVoucherPairsShipping []RuleVoucherPair
	for _, promotionRule := range shippingPriceRules {
		rule := &PriceRule{}
		*rule = promotionRule
		ruleVoucherPairsShipping = append(ruleVoucherPairsShipping, RuleVoucherPair{Rule: rule, Voucher: nil})
	}

	//apply them
	for _, priceRulePair := range ruleVoucherPairsShipping {
		orderDiscounts = calculateRule(orderDiscounts, priceRulePair, calculationParameters)
	}

	// ----------------------------------------------------------------------------------------------------------

	timeTrack(nowAll, "All rules together")
	for _, orderDiscount := range orderDiscounts {
		summary.TotalDiscount += orderDiscount.TotalDiscountAmount
		summary.TotalDiscountApplicable += orderDiscount.TotalDiscountAmountApplicable
		for _, appliedDiscount := range orderDiscount.AppliedDiscounts {
			summary.AppliedPriceRuleIDs = append(summary.AppliedPriceRuleIDs, appliedDiscount.PriceRuleID)
			if len(appliedDiscount.VoucherCode) > 0 {
				summary.AppliedVoucherCodes = append(summary.AppliedVoucherCodes, appliedDiscount.VoucherCode)
				summary.AppliedVoucherIDs = append(summary.AppliedVoucherIDs, appliedDiscount.VoucherID)
				if Verbose {
					log.Println("#### voucherCode: ", appliedDiscount.VoucherCode)
					log.Println("#### voucherID: ", appliedDiscount.VoucherID)
				}
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

	summary.TotalDiscountPercentage = summary.TotalDiscount / getOrderTotal(articleCollection, []string{}) * 100.0
	summary.TotalDiscountApplicablePercentage = summary.TotalDiscountApplicable / getOrderTotal(articleCollection, []string{}) * 100.0
	summary.AppliedPriceRuleIDs = RemoveDuplicates(summary.AppliedPriceRuleIDs)
	summary.AppliedVoucherIDs = RemoveDuplicates(summary.AppliedVoucherIDs)
	summary.AppliedVoucherCodes = RemoveDuplicates(summary.AppliedVoucherCodes)
	return orderDiscounts, summary, nil
}

// ApplyDiscountsOnCatalog applies all possible discounts on articleCollection ... if voucherCodes is "" the voucher is not applied
// This is not yet used. ApplyDiscounts should at some point be able to consider previousle calculated discounts
func ApplyDiscountsOnCatalog(articleCollection *ArticleCollection, existingDiscounts OrderDiscounts, roundTo float64, customProvider PriceRuleCustomProvider) (OrderDiscounts, *OrderDiscountSummary, error) {
	var ruleVoucherPairs []RuleVoucherPair
	start := time.Now()
	calculationParameters := &CalculationParameters{}
	calculationParameters.articleCollection = articleCollection
	calculationParameters.roundTo = roundTo
	calculationParameters.isCatalogCalculation = true
	calculationParameters.checkoutAttributes = []string{}

	now := time.Now()
	//find the groupIds for articleCollection items
	calculationParameters.productGroupIDsPerPosition = getProductGroupIDsPerPosition(articleCollection, true)
	timeTrack(now, "[ApplyDiscountsOnCatalog] loading of productGroupIDsPerPosition took ")
	now = time.Now()
	//find groups for customer
	groupIDsForCustomer := cache.GetGroupsIDSForItem(articleCollection.CustomerType, CustomerGroup)
	if len(groupIDsForCustomer) == 0 {
		groupIDsForCustomer = []string{}
	}
	calculationParameters.groupIDsForCustomer = groupIDsForCustomer
	timeTrack(now, "[ApplyDiscountsOnCatalog] loading of groupIDsForCustomer took ")

	calculationParameters.blacklistedItemIDs = cache.GetBlacklistedItemIDs()
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

	bestOptionCustomerProductRulePerItem := getBestOptionCustomerProductRulePerItem(ruleVoucherPairs, calculationParameters)
	calculationParameters.bestOptionCustomeProductRulePerItem = bestOptionCustomerProductRulePerItem
	//first loop where all promotion discounts are applied
	now = time.Now()
	for _, priceRulePair := range ruleVoucherPairs {
		pair := RuleVoucherPair{}
		pair = priceRulePair
		//apply them
		orderDiscounts = calculateRule(orderDiscounts, pair, calculationParameters)
	}

	timeTrack(now, "[ApplyDiscountsOnCatalog] CALCULATIONS and CHECKS took ")
	for _, orderDiscount := range orderDiscounts {
		for _, appliedDiscount := range orderDiscount.AppliedDiscounts {
			summary.AppliedPriceRuleIDs = append(summary.AppliedPriceRuleIDs, appliedDiscount.PriceRuleID)
		}
	}

	summary.AppliedPriceRuleIDs = RemoveDuplicates(summary.AppliedPriceRuleIDs)
	summary.AppliedVoucherIDs = []string{}
	summary.AppliedVoucherCodes = []string{}
	timeTrack(start, "[ApplyDiscountsOnCatalog] ALL DONE took ")
	return orderDiscounts, summary, nil
}

func calculateRule(orderDiscounts OrderDiscounts, priceRulePair RuleVoucherPair, calculationParameters *CalculationParameters) OrderDiscounts {
	ok := true
	if calculationParameters.isCatalogCalculation == false {
		ok, _ = validatePriceRuleForOrder(*priceRulePair.Rule, calculationParameters, orderDiscounts)
	} else {
		//bypass order check if catalog computation
		ok = true
	}

	if ok == true {
		switch priceRulePair.Rule.Action {
		case ActionItemByAbsolute:
			orderDiscounts = calculateDiscountsItemByAbsolute(priceRulePair, orderDiscounts, calculationParameters)
		case ActionItemByPercent:
			orderDiscounts = calculateDiscountsItemByPercent(priceRulePair, orderDiscounts, calculationParameters)
		case ActionCartByAbsolute:
			orderDiscounts = calculateDiscountsCartByAbsolute(priceRulePair, orderDiscounts, calculationParameters)
		case ActionCartByPercent:
			orderDiscounts = calculateDiscountsCartByPercentage(priceRulePair, orderDiscounts, calculationParameters)
		case ActionBuyXPayY:
			orderDiscounts = calculateDiscountsBuyXPayY(priceRulePair, orderDiscounts, calculationParameters)
		case ActionScaled:
			orderDiscounts = calculateScaledDiscounts(priceRulePair, orderDiscounts, calculationParameters)
		case ActionItemSetAbsolute:
			orderDiscounts = calculateItemSetAbsoluteDiscount(priceRulePair, orderDiscounts, calculationParameters)
		}
	}
	return orderDiscounts
}

// find what is the articleCollection value of positions that belong to group
// previouslyAppliedDiscounts is for qty 1
func getOrderTotalForPriceRule(priceRule *PriceRule, calculationParameters *CalculationParameters, orderDiscounts OrderDiscounts) float64 {
	var total float64

	for _, article := range calculationParameters.articleCollection.Articles {
		productGroupIDs := calculationParameters.productGroupIDsPerPosition[article.ID]
		if contains(article.ID, priceRule.ExcludedItemIDsFromOrderAmountCalculation) {
			continue
		}
		// rule has no customer or product group limitations
		if len(priceRule.IncludedProductGroupIDS) == 0 &&
			len(priceRule.ExcludedProductGroupIDS) == 0 &&
			len(priceRule.IncludedCustomerGroupIDS) == 0 &&
			len(priceRule.ExcludedCustomerGroupIDS) == 0 {
			total += article.Price * article.Quantity

			if priceRule.CalculateDiscountedOrderAmount == true {
				if orderDiscount, ok := orderDiscounts[article.ID]; ok {
					itemDiscount := orderDiscount.TotalDiscountAmountApplicable
					total = total - itemDiscount
				}
			}

		} else {
			//only sum up if limitations are matched
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, productGroupIDs) &&
				IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, productGroupIDs) &&
				IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) &&
				IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) {

				if priceRule.CalculateDiscountedOrderAmount == true {
					if orderDiscount, ok := orderDiscounts[article.ID]; ok {
						itemDiscount := orderDiscount.TotalDiscountAmountApplicable
						total = total - itemDiscount
					}
				}
				//--------------------------------------------------
				/*sub := 0.0
				if orderDiscounts != nil {
					previouslyAppliedDiscounts, ok := orderDiscounts[article.ID]
					if ok {
						sub = previouslyAppliedDiscounts.TotalDiscountAmount
					}
				}

				total += article.Price*article.Quantity - sub
				*/
			}
		}
	}
	return total
}

// find what is the articleCollection value of positions that belong to group
func getOrderTotal(articleCollection *ArticleCollection, excludedItemIDsFromOrderAmountCalculation []string) float64 {
	var total float64
	for _, article := range articleCollection.Articles {
		if contains(article.ID, excludedItemIDsFromOrderAmountCalculation) {
			continue
		}
		total += article.Price * article.Quantity
	}
	return total
}

// find what is the articleCollection total qty of positions that belong to group
// previouslyAppliedDiscounts is for qty 1
func getTotalQuantityForPriceRule(priceRule *PriceRule, calculationParameters *CalculationParameters) float64 {
	var total float64

	for _, article := range calculationParameters.articleCollection.Articles {
		productGroupIDs := calculationParameters.productGroupIDsPerPosition[article.ID]

		// rule has no customer or product group limitations
		if len(priceRule.IncludedProductGroupIDS) == 0 &&
			len(priceRule.ExcludedProductGroupIDS) == 0 &&
			len(priceRule.IncludedCustomerGroupIDS) == 0 &&
			len(priceRule.ExcludedCustomerGroupIDS) == 0 {
			total += article.Quantity
		} else {
			//only sum up if limitations are matched
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, productGroupIDs) &&
				IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, productGroupIDs) &&
				IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) &&
				IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) {

				total += article.Quantity
			}
		}
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
func validatePriceRuleForOrder(priceRule PriceRule, calculationParameters *CalculationParameters, orderDiscounts OrderDiscounts) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, nil, calculationParameters, orderDiscounts)
}

// ValidatePriceRuleForPosition -
func validatePriceRuleForPosition(priceRule PriceRule, article *Article, calculationParameters *CalculationParameters, orderDiscounts OrderDiscounts) (ok bool, reason TypeRuleValidationMsg) {
	return validatePriceRule(priceRule, article, calculationParameters, orderDiscounts)
}

// validatePriceRule -
func validatePriceRule(priceRule PriceRule, checkedPosition *Article, calculationParameters *CalculationParameters, orderDiscounts OrderDiscounts) (ok bool, reason TypeRuleValidationMsg) {
	if len(priceRule.CheckoutAttributes) > 0 {
		match := false
		for _, checkoutAttribute := range calculationParameters.checkoutAttributes {
			if contains(checkoutAttribute, priceRule.CheckoutAttributes) {
				match = true
				break
			}
		}
		if !match {
			return false, ValidationPriceRuleCheckoutAttributesMismatch
		}
	}

	if calculationParameters.isCatalogCalculation {
		if priceRule.QtyThreshold > 0 || priceRule.MinOrderAmount > 0 {
			return false, ValidationPriceRuleNotForCatalogueCalculation
		}
	}

	if !calculationParameters.isCatalogCalculation {
		if priceRule.MaxUses <= priceRule.UsageHistory.TotalUsages {
			return false, ValidationPriceRuleMaxUsages
		}
		// if we have the use and the customer history usage ... check ...
		if customerUsages, ok := priceRule.UsageHistory.UsagesPerCustomer[calculationParameters.articleCollection.CustomerID]; ok && len(calculationParameters.articleCollection.CustomerID) > 0 {
			if priceRule.MaxUsesPerCustomer <= customerUsages {
				return false, ValidationPriceRuleMaxUsagesPerCustomer
			}
		}

		if priceRule.MinOrderAmount > 0.0 {
			if priceRule.MinOrderAmountApplicableItemsOnly {
				if priceRule.MinOrderAmount > getOrderTotalForPriceRule(&priceRule, calculationParameters, orderDiscounts) {
					return false, ValidationPriceRuleMinimumAmount
				}
			} else {
				orderTotal := getOrderTotal(calculationParameters.articleCollection, priceRule.ExcludedItemIDsFromOrderAmountCalculation)

				//remove previously discounted amount if rule config says so
				if priceRule.CalculateDiscountedOrderAmount {
					for itemID, dicountCalculationData := range orderDiscounts {
						if !contains(itemID, priceRule.ExcludedItemIDsFromOrderAmountCalculation) {
							orderTotal -= dicountCalculationData.TotalDiscountAmount
						}
					}
				}

				if priceRule.MinOrderAmount > orderTotal {
					return false, ValidationPriceRuleMinimumAmount
				}
			}

		}

		//check the pricerule qty threshold
		if priceRule.QtyThreshold > 0 {
			ruleQtyThreshold := getTotalQuantityForPriceRule(&priceRule, calculationParameters)
			if ruleQtyThreshold < priceRule.QtyThreshold {
				return false, ValidationPriceRuleQuantityThresholdNotMet
			}
		}
	}

	now := time.Now()
	var productGroupIncludeMatchOK = false
	var productGroupExcludeMatchOK = false
	var customerGroupIncludeMatchOK = false
	var customerGroupExcludeMatchOK = false
	var blacklistOK = false

	if checkedPosition == nil {
		for _, article := range calculationParameters.articleCollection.Articles {
			//at least some must match
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, calculationParameters.productGroupIDsPerPosition[article.ID]) {
				productGroupIncludeMatchOK = true
			}
			if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, calculationParameters.productGroupIDsPerPosition[article.ID]) {
				productGroupExcludeMatchOK = true
			}
			if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) {
				customerGroupIncludeMatchOK = true
			}
			if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) {
				customerGroupExcludeMatchOK = true
			}
			if !contains(article.ID, calculationParameters.blacklistedItemIDs) {
				blacklistOK = true
			}
		}
	} else {
		// if only checking for one item, do not go through loop
		if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedProductGroupIDS, calculationParameters.productGroupIDsPerPosition[checkedPosition.ID]) {
			productGroupIncludeMatchOK = true
		}
		if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedProductGroupIDS, calculationParameters.productGroupIDsPerPosition[checkedPosition.ID]) {
			productGroupExcludeMatchOK = true
		}
		if IsOneProductOrCustomerGroupInIncludedGroups(priceRule.IncludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) {
			customerGroupIncludeMatchOK = true
		}

		if IsNoProductOrGroupInExcludeGroups(priceRule.ExcludedCustomerGroupIDS, calculationParameters.groupIDsForCustomer) {
			customerGroupExcludeMatchOK = true
		}

		if !contains(checkedPosition.ID, calculationParameters.blacklistedItemIDs) {
			blacklistOK = true
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

	if !blacklistOK {
		return false, ValidationPriceRuleBlacklist
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
	if Verbose {
		elapsed := time.Since(start)
		log.Printf("%s %s", name, elapsed)
	}
}

func contains(e string, s []string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getBestOptionCustomerProductRulePerItem(ruleVoucherPairs []RuleVoucherPair, calculationParameters *CalculationParameters) (ret map[string]string) {
	start := time.Now()
	ret = make(map[string]string)
	currentDiscounts := make(map[string]float64)
	currentBestDiscountType := make(map[string]Type)

	for _, priceRulePair := range ruleVoucherPairs {
		tempDiscounts := NewOrderDiscounts(calculationParameters.articleCollection)
		tempDiscounts = calculateRule(tempDiscounts, priceRulePair, calculationParameters)
		//go over applied discounts an select the better one
		for _, article := range calculationParameters.articleCollection.Articles {
			itemID := article.ID
			discount := tempDiscounts[itemID].TotalDiscountAmount
			//init map if necessary
			if _, ok := currentDiscounts[itemID]; !ok {
				currentDiscounts[itemID] = 0
			}
			//init map if necessary
			if _, ok := currentBestDiscountType[itemID]; !ok {
				currentBestDiscountType[itemID] = TypePromotionCustomer // we can always overrider
			}

			if (discount > currentDiscounts[itemID] && currentBestDiscountType[itemID] == TypePromotionCustomer) ||
				(discount > currentDiscounts[itemID] && currentBestDiscountType[itemID] == TypePromotionProduct && priceRulePair.Rule.Type != TypePromotionCustomer) {
				currentDiscounts[itemID] = discount
				currentBestDiscountType[itemID] = priceRulePair.Rule.Type
				ret[itemID] = tempDiscounts[itemID].AppliedDiscounts[0].PriceRuleID
			}
		}
	}
	timeTrack(start, "getBestOptionCustomerProductRulePerItem took")
	return ret
}
