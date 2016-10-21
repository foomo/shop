package pricerule

import "github.com/foomo/shop/order"

// CalculateDiscountsCartByPercentage -
func calculateDiscountsCartByPercentage(order *order.Order, priceRuleVoucherPair *RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionCartByPercent {
		panic("CalculateDiscountsCartByPercentage called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	orderTotal := getOrderTotalForPriceRule(priceRuleVoucherPair.Rule, order, productGroupIDsPerPosition, groupIDsForCustomer)
	//the discount amount calculation
	totalDiscountAmount := roundToStep(orderTotal*priceRuleVoucherPair.Rule.Amount/100.0, roundTo)

	//from here we call existing methods with a hacked priceRule that will keep the name and ID but different action and amount
	tempPriceRule := *priceRuleVoucherPair.Rule
	tempPriceRule.Action = ActionCartByAbsolute
	tempPriceRule.Amount = totalDiscountAmount
	tempPriceRuleVoucherPair := RuleVoucherPair{Rule: &tempPriceRule, Voucher: priceRuleVoucherPair.Voucher}

	return calculateDiscountsCartByAbsolute(order, &tempPriceRuleVoucherPair, orderDiscounts, productGroupIDsPerPosition, groupIDsForCustomer, roundTo)
}
