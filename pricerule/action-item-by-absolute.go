package pricerule

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByAbsolute(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByAbsolute {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, article := range calculationParameters.articleCollection.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, article, calculationParameters, orderDiscounts)

		orderDiscountsForPosition := orderDiscounts[article.ID]
		if !orderDiscounts[article.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition, calculationParameters) {
			//apply the discount here
			discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, article.ID)

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep((orderDiscounts[article.ID].Quantity * priceRuleVoucherPair.Rule.Amount), calculationParameters.roundTo)
			discountApplied.DiscountSingle = priceRuleVoucherPair.Rule.Amount
			discountApplied.Quantity = orderDiscounts[article.ID].Quantity

			//pointer assignment WTF !!!
			orderDiscountsForPosition := orderDiscounts[article.ID]
			orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, article.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, calculationParameters.roundTo)
			orderDiscounts[article.ID] = orderDiscountsForPosition
		}
	}
	return orderDiscounts
}

//allow only one customer or product promotion ... allow only the best one if specified
func previouslyAppliedExclusionInPlace(rule *PriceRule, orderDiscountsForPosition DiscountCalculationData, calculationParameters *CalculationParameters) bool {
	itemID := orderDiscountsForPosition.OrderItemID
	previouslyAppliedExclusion := false

	if rule.Type == TypePromotionCustomer || rule.Type == TypePromotionProduct {
		if calculationParameters.bestOptionCustomeProductRulePerItem != nil {
			if bestRuleID, ok := calculationParameters.bestOptionCustomeProductRulePerItem[itemID]; ok {
				if rule.ID == bestRuleID {
					return false
				}
				return true
			}
		}

		if orderDiscountsForPosition.CustomerPromotionApplied || orderDiscountsForPosition.ProductPromotionApplied {
			previouslyAppliedExclusion = true
		}
	}
	return previouslyAppliedExclusion
}
