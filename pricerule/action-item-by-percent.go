package pricerule

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByPercent(itemCollection *ItemCollection, priceRuleVoucherPair RuleVoucherPair, itemCollDiscounts OrderDiscounts, productGroupIDsPerItem map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByPercent {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, item := range itemCollection.Items {
		ok, _ := validatePriceRuleForItem(*priceRuleVoucherPair.Rule, itemCollection, item, productGroupIDsPerItem, groupIDsForCustomer)

		itemCollDiscountsForItem := itemCollDiscounts[item.ID]
		if !itemCollDiscounts[item.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, itemCollDiscountsForItem) {
			//apply the discount here
			discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, itemCollDiscounts, item.ID)

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep(priceRuleVoucherPair.Rule.Amount/100*discountApplied.CalculationBasePrice*itemCollDiscounts[item.ID].Quantity, roundTo)
			discountApplied.DiscountSingle = roundToStep(priceRuleVoucherPair.Rule.Amount/100*discountApplied.CalculationBasePrice, roundTo)
			discountApplied.Quantity = itemCollDiscounts[item.ID].Quantity

			itemCollDiscountsForItem := itemCollDiscounts[item.ID]
			itemCollDiscountsForItem = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, item.ID, itemCollDiscountsForItem, itemCollDiscounts, *priceRuleVoucherPair.Rule, roundTo)

			itemCollDiscounts[item.ID] = itemCollDiscountsForItem
		}
	}
	return itemCollDiscounts
}

func getInitializedDiscountApplied(priceRuleVoucherPair RuleVoucherPair, itemCollDiscounts OrderDiscounts, itemID string) *DiscountApplied {
	discountApplied := &DiscountApplied{}
	discountApplied.PriceRuleID = priceRuleVoucherPair.Rule.ID
	discountApplied.MappingID = priceRuleVoucherPair.Rule.MappingID
	if priceRuleVoucherPair.Voucher != nil {
		discountApplied.VoucherID = priceRuleVoucherPair.Voucher.ID
		discountApplied.VoucherCode = priceRuleVoucherPair.Voucher.VoucherCode
	}
	if priceRuleVoucherPair.Rule.Type != TypeVoucher {
		discountApplied.CalculationBasePrice = itemCollDiscounts[itemID].InitialItemPrice
	} else {
		discountApplied.CalculationBasePrice = itemCollDiscounts[itemID].VoucherCalculationBaseItemPrice
	}
	discountApplied.Price = itemCollDiscounts[itemID].InitialItemPrice
	return discountApplied
}

func calculateCurrentPriceAndApplicableDiscountsEnforceRules(discountApplied DiscountApplied, itemID string, itemCollDiscountsForItem DiscountCalculationData, itemCollDiscounts OrderDiscounts, rule PriceRule, roundTo float64) DiscountCalculationData {
	// make sure the discount is not more that can be actually given ... discount < price
	if itemCollDiscountsForItem.CurrentItemPrice < discountApplied.DiscountSingle {
		discountApplied.DiscountSingleApplicable = roundToStep(itemCollDiscountsForItem.CurrentItemPrice, roundTo)
		discountApplied.DiscountAmountApplicable = discountApplied.DiscountSingleApplicable * itemCollDiscounts[itemID].Quantity

	} else {
		discountApplied.DiscountSingleApplicable = discountApplied.DiscountSingle
		discountApplied.DiscountAmountApplicable = discountApplied.DiscountAmount
	}

	itemCollDiscountsForItem.CurrentItemPrice = itemCollDiscountsForItem.CurrentItemPrice - discountApplied.DiscountSingleApplicable
	itemCollDiscountsForItem.TotalDiscountAmount += discountApplied.DiscountAmount
	itemCollDiscountsForItem.TotalDiscountAmountApplicable += discountApplied.DiscountAmountApplicable

	//store the reduced price so that it will be used for the vouchers calculation
	if rule.Type != TypeVoucher {
		itemCollDiscountsForItem.VoucherCalculationBaseItemPrice = itemCollDiscountsForItem.CurrentItemPrice
	}

	if rule.Exclusive {
		itemCollDiscountsForItem.StopApplyingDiscounts = true
	}
	//mark the type applied - see method previouslyAppliedExclusionInPlace
	if rule.Type == TypePromotionCustomer {
		itemCollDiscountsForItem.CustomerPromotionApplied = true
	}
	if rule.Type == TypePromotionProduct {
		itemCollDiscountsForItem.ProductPromotionApplied = true
	}
	itemCollDiscountsForItem.AppliedDiscounts = append(itemCollDiscountsForItem.AppliedDiscounts, discountApplied)

	return itemCollDiscountsForItem
}
