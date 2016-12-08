package pricerule

import "github.com/foomo/shop/order"

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByPercent(order *order.Order, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByPercent {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, position := range order.Positions {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, order, position, productGroupIDsPerPosition, groupIDsForCustomer)

		orderDiscountsForPosition := orderDiscounts[position.ItemID]
		if !orderDiscounts[position.ItemID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition) {
			//apply the discount here
			discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, position.ItemID)

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep(priceRuleVoucherPair.Rule.Amount/100*discountApplied.CalculationBasePrice*orderDiscounts[position.ItemID].Quantity, roundTo)
			discountApplied.DiscountSingle = roundToStep(priceRuleVoucherPair.Rule.Amount/100*discountApplied.CalculationBasePrice, roundTo)
			discountApplied.Quantity = orderDiscounts[position.ItemID].Quantity

			orderDiscountsForPosition := orderDiscounts[position.ItemID]
			orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, position.ItemID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, roundTo)

			orderDiscounts[position.ItemID] = orderDiscountsForPosition
		}
	}
	return orderDiscounts
}

func getInitializedDiscountApplied(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, itemID string) *DiscountApplied {
	discountApplied := &DiscountApplied{}
	discountApplied.PriceRuleID = priceRuleVoucherPair.Rule.ID
	discountApplied.MappingID = priceRuleVoucherPair.Rule.MappingID
	if priceRuleVoucherPair.Voucher != nil {
		discountApplied.VoucherID = priceRuleVoucherPair.Voucher.ID
		discountApplied.VoucherCode = priceRuleVoucherPair.Voucher.VoucherCode
	}
	if priceRuleVoucherPair.Rule.Type != TypeVoucher {
		discountApplied.CalculationBasePrice = orderDiscounts[itemID].InitialItemPrice
	} else {
		discountApplied.CalculationBasePrice = orderDiscounts[itemID].VoucherCalculationBaseItemPrice
	}
	discountApplied.Price = orderDiscounts[itemID].InitialItemPrice
	return discountApplied
}

func calculateCurrentPriceAndApplicableDiscountsEnforceRules(discountApplied DiscountApplied, itemID string, orderDiscountsForPosition DiscountCalculationData, orderDiscounts OrderDiscounts, rule PriceRule, roundTo float64) DiscountCalculationData {
	// make sure the discount is not more that can be actually given ... discount < price
	if orderDiscountsForPosition.CurrentItemPrice < discountApplied.DiscountSingle {
		discountApplied.DiscountSingleApplicable = roundToStep(orderDiscountsForPosition.CurrentItemPrice, roundTo)
		discountApplied.DiscountAmountApplicable = discountApplied.DiscountSingleApplicable * orderDiscounts[itemID].Quantity

	} else {
		discountApplied.DiscountSingleApplicable = discountApplied.DiscountSingle
		discountApplied.DiscountAmountApplicable = discountApplied.DiscountAmount
	}

	orderDiscountsForPosition.CurrentItemPrice = orderDiscountsForPosition.CurrentItemPrice - discountApplied.DiscountSingleApplicable
	orderDiscountsForPosition.TotalDiscountAmount += discountApplied.DiscountAmount
	orderDiscountsForPosition.TotalDiscountAmountApplicable += discountApplied.DiscountAmountApplicable

	//store the reduced price so that it will be used for the vouchers calculation
	if rule.Type != TypeVoucher {
		orderDiscountsForPosition.VoucherCalculationBaseItemPrice = orderDiscountsForPosition.CurrentItemPrice
	}

	if rule.Exclusive {
		orderDiscountsForPosition.StopApplyingDiscounts = true
	}
	//mark the type applied - see method previouslyAppliedExclusionInPlace
	if rule.Type == TypePromotionCustomer {
		orderDiscountsForPosition.CustomerPromotionApplied = true
	}
	if rule.Type == TypePromotionProduct {
		orderDiscountsForPosition.ProductPromotionApplied = true
	}
	orderDiscountsForPosition.AppliedDiscounts = append(orderDiscountsForPosition.AppliedDiscounts, discountApplied)

	return orderDiscountsForPosition
}