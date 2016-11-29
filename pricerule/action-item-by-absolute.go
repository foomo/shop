package pricerule

import "github.com/foomo/shop/order"

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByAbsolute(order *order.Order, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByAbsolute {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, position := range order.Positions {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, order, position, productGroupIDsPerPosition, groupIDsForCustomer)

		orderDiscountsForPosition := orderDiscounts[position.ItemID]
		if !orderDiscounts[position.ItemID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition) {
			//apply the discount here
			discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, position.ItemID)

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep((orderDiscounts[position.ItemID].Quantity * priceRuleVoucherPair.Rule.Amount), roundTo)
			discountApplied.DiscountSingle = priceRuleVoucherPair.Rule.Amount
			discountApplied.Quantity = orderDiscounts[position.ItemID].Quantity

			//pointer assignment WTF !!!
			orderDiscountsForPosition := orderDiscounts[position.ItemID]
			orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, position.ItemID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, roundTo)
			orderDiscounts[position.ItemID] = orderDiscountsForPosition
		}
	}
	return orderDiscounts
}

func previouslyAppliedExclusionInPlace(rule *PriceRule, orderDiscountsForPosition DiscountCalculationData) bool {
	previouslyAppliedExclusion := false
	if rule.Type == TypePromotionCustomer || rule.Type == TypePromotionProduct {
		if orderDiscountsForPosition.CustomerPromotionApplied || orderDiscountsForPosition.ProductPromotionApplied {
			previouslyAppliedExclusion = true
		}
	}
	return previouslyAppliedExclusion
}
