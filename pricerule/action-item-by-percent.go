package pricerule

import "github.com/foomo/shop/order"

// CalculateDiscountsItemByPercent -
func calculateDiscountsItemByPercent(order *order.Order, priceRuleVoucherPair *RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemByPercent {
		panic("CalculateDiscountsItemByPercent called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	for _, position := range order.Positions {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, order, position, productGroupIDsPerPosition, groupIDsForCustomer)

		if !orderDiscounts[position.ItemID].StopApplyingDiscounts && ok {
			//apply the discount here
			discountApplied := &DiscountApplied{}
			discountApplied.PriceRuleID = priceRuleVoucherPair.Rule.ID
			discountApplied.MappingID = priceRuleVoucherPair.Rule.MappingID
			if priceRuleVoucherPair.Voucher != nil {
				discountApplied.VoucherID = priceRuleVoucherPair.Voucher.ID
				discountApplied.VoucherCode = priceRuleVoucherPair.Voucher.VoucherCode
			}
			discountApplied.CalculationBasePrice = orderDiscounts[position.ItemID].CurrentItemPrice
			discountApplied.Price = orderDiscounts[position.ItemID].InitialItemPrice

			//calculate the actual discount
			discountApplied.DiscountAmount = roundToStep(priceRuleVoucherPair.Rule.Amount/100*discountApplied.CalculationBasePrice*orderDiscounts[position.ItemID].Qantity, roundTo)
			discountApplied.DiscountSingle = roundToStep(priceRuleVoucherPair.Rule.Amount/100*discountApplied.CalculationBasePrice, roundTo)
			discountApplied.Quantity = orderDiscounts[position.ItemID].Qantity

			//pointer assignment WTF !!!
			orderDiscountsForPosition := orderDiscounts[position.ItemID]
			orderDiscountsForPosition.TotalDiscountAmount += discountApplied.DiscountAmount
			orderDiscountsForPosition.AppliedDiscounts = append(orderDiscountsForPosition.AppliedDiscounts, *discountApplied)
			if priceRuleVoucherPair.Rule.Exclusive {
				orderDiscountsForPosition.StopApplyingDiscounts = true
			}
			orderDiscounts[position.ItemID] = orderDiscountsForPosition
		}
	}
	return orderDiscounts
}
