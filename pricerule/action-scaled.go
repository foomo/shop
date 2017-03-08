package pricerule

import "log"

// CalculateScaledDiscounts -
func calculateScaledDiscounts(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionScaled {
		panic("CalculateScaledDiscounts called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	if calculationParameters.isCatalogCalculation == true {
		if Verbose {
			log.Println("catalog calculations can not handle actions of type CalculateScaledDiscounts")
		}
		return orderDiscounts
	}
	orderTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, calculationParameters, orderDiscounts)
	totalQuantityForRule := getTotalQuantityForRule(*priceRuleVoucherPair.Rule, calculationParameters, orderDiscounts)

	//check if we have a matching scale
	//note: the first matching is picked -> scales must not overlap
	for _, scaledLevel := range priceRuleVoucherPair.Rule.ScaledAmounts {
		compareValue := orderTotal
		if !scaledLevel.IsFromToPrice {
			compareValue = totalQuantityForRule
		}
		if compareValue >= scaledLevel.FromValue && compareValue <= scaledLevel.ToValue {
			//from here we call existing methods with a temp price rule pair will keep the name and ID but different action and amount
			return evaluateScale(scaledLevel, priceRuleVoucherPair, orderDiscounts, calculationParameters)
		}
	}
	return orderDiscounts
}

func evaluateScale(scaledLevel ScaledAmountLevel, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {
	scaledPriceRule := *priceRuleVoucherPair.Rule
	scaledPriceRule.Amount = scaledLevel.Amount
	if scaledLevel.IsScaledAmountPercentage {
		scaledPriceRule.Action = ActionCartByPercent
		tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}
		return calculateDiscountsCartByPercentage(tempRuleVoucherPair, orderDiscounts, calculationParameters)
	}
	scaledPriceRule.Action = ActionCartByAbsolute
	tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}
	return calculateDiscountsCartByAbsolute(tempRuleVoucherPair, orderDiscounts, calculationParameters)
}

func getTotalQuantityForRule(priceRule PriceRule, calculationParameters *CalculationParameters, orderDiscounts OrderDiscounts) float64 {
	totalQty := 0.0
	for _, article := range calculationParameters.articleCollection.Articles {
		if ok, _ := validatePriceRule(priceRule, article, calculationParameters, orderDiscounts); ok {
			totalQty += article.Quantity
		}
	}
	return totalQty
}
