package pricerule

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByAbsolute(articleCollection *ArticleCollection, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByAbsolute {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, article := range articleCollection.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, articleCollection, article, productGroupIDsPerPosition, groupIDsForCustomer)

		orderDiscountsForPosition := orderDiscounts[article.ID]
		if !orderDiscounts[article.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition) {
			//apply the discount here
			discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, article.ID)

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep((orderDiscounts[article.ID].Quantity * priceRuleVoucherPair.Rule.Amount), roundTo)
			discountApplied.DiscountSingle = priceRuleVoucherPair.Rule.Amount
			discountApplied.Quantity = orderDiscounts[article.ID].Quantity

			//pointer assignment WTF !!!
			orderDiscountsForPosition := orderDiscounts[article.ID]
			orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, article.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, roundTo)
			orderDiscounts[article.ID] = orderDiscountsForPosition
		}
	}
	return orderDiscounts
}

func previouslyAppliedExclusionInPlace(rule *PriceRule, orderDiscountsForPosition DiscountCalculationData) bool {
	previouslyAppliedExclusion := false
	// if rule.Type == TypePromotionProduct {
	// 	if orderDiscountsForPosition.ProductPromotionApplied {
	// 		previouslyAppliedExclusion = true
	// 	}
	// }
	if rule.Type == TypePromotionCustomer || rule.Type == TypePromotionProduct {
		if orderDiscountsForPosition.CustomerPromotionApplied || orderDiscountsForPosition.ProductPromotionApplied {
			previouslyAppliedExclusion = true
		}
	}
	return previouslyAppliedExclusion
}
