package pricerule

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByAbsolute(itemCollection *ItemCollection, priceRuleVoucherPair RuleVoucherPair, itemCollDiscounts OrderDiscounts, productGroupIDsPerItem map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByAbsolute {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, item := range itemCollection.Items {
		ok, _ := validatePriceRuleForItem(*priceRuleVoucherPair.Rule, itemCollection, item, productGroupIDsPerItem, groupIDsForCustomer)

		itemCollDiscountsForItem := itemCollDiscounts[item.ID]
		if !itemCollDiscounts[item.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, itemCollDiscountsForItem) {
			//apply the discount here
			discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, itemCollDiscounts, item.ID)

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep((itemCollDiscounts[item.ID].Quantity * priceRuleVoucherPair.Rule.Amount), roundTo)
			discountApplied.DiscountSingle = priceRuleVoucherPair.Rule.Amount
			discountApplied.Quantity = itemCollDiscounts[item.ID].Quantity

			//pointer assignment WTF !!!
			itemCollDiscountsForItem := itemCollDiscounts[item.ID]
			itemCollDiscountsForItem = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, item.ID, itemCollDiscountsForItem, itemCollDiscounts, *priceRuleVoucherPair.Rule, roundTo)
			itemCollDiscounts[item.ID] = itemCollDiscountsForItem
		}
	}
	return itemCollDiscounts
}

func previouslyAppliedExclusionInPlace(rule *PriceRule, itemCollDiscountsForItem DiscountCalculationData) bool {
	previouslyAppliedExclusion := false
	if rule.Type == TypePromotionCustomer || rule.Type == TypePromotionProduct {
		if itemCollDiscountsForItem.CustomerPromotionApplied || itemCollDiscountsForItem.ProductPromotionApplied {
			previouslyAppliedExclusion = true
		}
	}
	return previouslyAppliedExclusion
}
