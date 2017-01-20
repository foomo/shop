package pricerule

// CalculateScaledDiscounts -
func calculateScaledDiscounts(itemCollection *ItemCollection, priceRuleVoucherPair RuleVoucherPair, itemCollDiscounts OrderDiscounts, productGroupIDsPerItem map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionScaled {
		panic("CalculateScaledDiscounts called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	itemCollTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, itemCollection, productGroupIDsPerItem, groupIDsForCustomer)
	//check if we have a matching scale
	//note: the first matching is picked -> scales must not overlap
	for _, scaledLevel := range priceRuleVoucherPair.Rule.ScaledAmounts {
		if itemCollTotal >= scaledLevel.FromValue && itemCollTotal <= scaledLevel.ToValue {
			//from here we call existing methods with a temp price rule pair will keep the name and ID but different action and amount
			scaledPriceRule := *priceRuleVoucherPair.Rule
			scaledPriceRule.Amount = scaledLevel.Amount
			if scaledLevel.IsScaledAmountPercentage {
				scaledPriceRule.Action = ActionCartByPercent
				tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}

				return calculateDiscountsCartByPercentage(itemCollection, tempRuleVoucherPair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
			}

			scaledPriceRule.Action = ActionCartByAbsolute
			tempRuleVoucherPair := RuleVoucherPair{Rule: &scaledPriceRule, Voucher: priceRuleVoucherPair.Voucher}

			return calculateDiscountsCartByAbsolute(itemCollection, tempRuleVoucherPair, itemCollDiscounts, productGroupIDsPerItem, groupIDsForCustomer, roundTo)
		}
	}
	return itemCollDiscounts
}
