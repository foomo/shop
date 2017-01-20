package pricerule

// CalculateDiscountsCartByPercentage -
func calculateDiscountsCartByPercentage(itemCollection *ItemCollection, priceRuleVoucherPair RuleVoucherPair, itemCollDiscounts OrderDiscounts, productGroupIDsPerItem map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionCartByPercent {
		panic("CalculateDiscountsCartByPercentage called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	//get the total - for vouchers it is lowered by previous discounts
	itemCollTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, itemCollection, productGroupIDsPerItem, groupIDsForCustomer)

	//the discount amount calculation
	totalDiscountAmount := roundToStep(itemCollTotal*priceRuleVoucherPair.Rule.Amount/100.0, roundTo)

	//from here we call existing methods with a hacked priceRule that will keep the name and ID but different action and amount
	tempPriceRule := *priceRuleVoucherPair.Rule
	tempPriceRule.Action = ActionCartByAbsolute
	tempPriceRule.Amount = totalDiscountAmount
	tempPriceRuleVoucherPair := RuleVoucherPair{Rule: &tempPriceRule, Voucher: priceRuleVoucherPair.Voucher}

	return calculateDiscountsCartByAbsolute(itemCollection, tempPriceRuleVoucherPair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
}
