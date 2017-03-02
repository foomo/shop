package pricerule

import "log"

// CalculateDiscountsCartByPercentage -
func calculateDiscountsCartByPercentage(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionCartByPercent {
		panic("CalculateDiscountsCartByPercentage called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	if calculationParameters.isCatalogCalculation == true {
		if Verbose {
			log.Println("catalog calculations can not handle actions of type CalculateDiscountsCartByPercentage")
		}
		return orderDiscounts
	}

	//get the total - for vouchers it is lowered by previous discounts
	orderTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, calculationParameters, orderDiscounts)
	//the discount amount calculation
	totalDiscountAmount := roundToStep(orderTotal*priceRuleVoucherPair.Rule.Amount/100.0, calculationParameters.roundTo)

	//from here we call existing methods with a hacked priceRule that will keep the name and ID but different action and amount
	tempPriceRule := *priceRuleVoucherPair.Rule
	tempPriceRule.Action = ActionCartByAbsolute
	tempPriceRule.Amount = totalDiscountAmount
	tempPriceRuleVoucherPair := RuleVoucherPair{Rule: &tempPriceRule, Voucher: priceRuleVoucherPair.Voucher}

	return calculateDiscountsCartByAbsolute(tempPriceRuleVoucherPair, orderDiscounts, calculationParameters)
}
