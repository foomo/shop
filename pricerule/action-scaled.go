package pricerule

import "log"

// CalculateScaledDiscounts -
func calculateScaledDiscounts(articleCollection *ArticleCollection, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64, isCatalogCalculation bool) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionScaled {
		panic("CalculateScaledDiscounts called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	if isCatalogCalculation == true {
		log.Println("catalog calculations can not handle actions of type CalculateScaledDiscounts")
		return orderDiscounts
	}
	orderTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, orderDiscounts)
	totalQuantityForRule := getTotalQuantityForRule(*priceRuleVoucherPair.Rule, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, orderDiscounts, isCatalogCalculation)

	//check if we have a matching scale
	//note: the first matching is picked -> scales must not overlap
	for _, scaledLevel := range priceRuleVoucherPair.Rule.ScaledAmounts {
		compareValue := orderTotal
		if !scaledLevel.IsFromToPrice {
			compareValue = totalQuantityForRule
		}
		if compareValue >= scaledLevel.FromValue && compareValue <= scaledLevel.ToValue {
			//from here we call existing methods with a temp price rule pair will keep the name and ID but different action and amount
			return evaluateScale(scaledLevel, priceRuleVoucherPair, articleCollection, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
		}
	}
	return orderDiscounts
}

func evaluateScale(scaledLevel ScaledAmountLevel, priceRuleVoucherPair RuleVoucherPair, articleCollection *ArticleCollection, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64, isCatalogCalculation bool) OrderDiscounts {
	scaledPriceRule := *priceRuleVoucherPair.Rule
	scaledPriceRule.Amount = scaledLevel.Amount
	if scaledLevel.IsScaledAmountPercentage {
		scaledPriceRule.Action = ActionCartByPercent
		tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}
		return calculateDiscountsCartByPercentage(articleCollection, tempRuleVoucherPair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
	}
	scaledPriceRule.Action = ActionCartByAbsolute
	tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}
	return calculateDiscountsCartByAbsolute(articleCollection, tempRuleVoucherPair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo, isCatalogCalculation)
}

func getTotalQuantityForRule(priceRule PriceRule, articleCollection *ArticleCollection, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, orderDiscounts OrderDiscounts, isCatalogCalculation bool) float64 {
	totalQty := 0.0
	for _, article := range articleCollection.Articles {
		if ok, _ := validatePriceRule(priceRule, articleCollection, article, productGroupIDsPerPosition, groupIDsForCustomer, isCatalogCalculation); ok {
			totalQty += article.Quantity
		}
	}

	return totalQty
}
