package pricerule

import (
	"log"
)

// CalculateDiscountsItemByPercent -
func calculateItemSetAbsoluteDiscount(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {

	if priceRuleVoucherPair.Rule.Action != ActionItemSetAbsolute {
		panic("CalculateItemSetAbsoluteDiscount called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	if calculationParameters.isCatalogCalculation == true {
		log.Println("catalog calculations can not handle actions of type ActionItemSetAbsolute")
		return orderDiscounts
	}

	_, timesSetIncluded := getOrderItemsThatBelongToSet(priceRuleVoucherPair.Rule, calculationParameters.articleCollection)
	if timesSetIncluded >= 1.0 {
		cartDiscountAmount := priceRuleVoucherPair.Rule.Amount * timesSetIncluded
		tempPriceRule := *priceRuleVoucherPair.Rule
		tempPriceRule.Amount = cartDiscountAmount
		tempPriceRule.Action = ActionCartByAbsolute
		tempPriceRule.IsAmountIndependentOfQty = priceRuleVoucherPair.Rule.IsAmountIndependentOfQty
		log.Println("rule action ActionItemSetAbsolute internally converted to ActionCartAbsolute")
		priceRuleVoucherPair.Rule = &tempPriceRule
		return calculateDiscountsCartByAbsolute(priceRuleVoucherPair, orderDiscounts, calculationParameters)
	}
	log.Println("item set present ", timesSetIncluded, " times")
	return orderDiscounts
}

// how many times is the whole set present
func getOrderItemsThatBelongToSet(priceRule *PriceRule, articleCollection *ArticleCollection) (itemIDs []string, timesSetIncluded float64) {
	itemIDs = []string{}
	countItemsForSetIndex := make(map[int]float64)
	for _, article := range articleCollection.Articles {
		//if article.ID is in one of the itemset arrays
		for setIndex, set := range priceRule.ItemSets {
			if _, ok := countItemsForSetIndex[setIndex]; !ok {
				countItemsForSetIndex[setIndex] = 0.0
			}
			if contains(article.ID, set) {
				itemIDs = append(itemIDs, article.ID)
				countItemsForSetIndex[setIndex] += article.Quantity
			}
		}
	}

	timesSetIncluded = 0.0
	for _, count := range countItemsForSetIndex {
		if timesSetIncluded == 0 {
			timesSetIncluded = count
		}
		if count < timesSetIncluded {
			timesSetIncluded = count
		}
	}
	return itemIDs, timesSetIncluded
}
