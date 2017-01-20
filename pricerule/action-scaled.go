package pricerule

// CalculateScaledDiscounts -
func calculateScaledDiscounts(articleCollection *ArticleCollection, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionScaled {
		panic("CalculateScaledDiscounts called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	orderTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer)
	//check if we have a matching scale
	//note: the first matching is picked -> scales must not overlap
	for _, scaledLevel := range priceRuleVoucherPair.Rule.ScaledAmounts {
		if orderTotal >= scaledLevel.FromValue && orderTotal <= scaledLevel.ToValue {
			//from here we call existing methods with a temp price rule pair will keep the name and ID but different action and amount
			scaledPriceRule := *priceRuleVoucherPair.Rule
			scaledPriceRule.Amount = scaledLevel.Amount
			if scaledLevel.IsScaledAmountPercentage {
				scaledPriceRule.Action = ActionCartByPercent
				tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}

				return calculateDiscountsCartByPercentage(articleCollection, tempRuleVoucherPair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
			}

			scaledPriceRule.Action = ActionCartByAbsolute
			tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}

			return calculateDiscountsCartByAbsolute(articleCollection, tempRuleVoucherPair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
		}
	}
	return orderDiscounts
}
