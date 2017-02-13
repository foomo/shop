package pricerule

import (
	"log"
	"math"
	"sort"
)

// CalculateDiscountsBuyXGetY -
func calculateDiscountsBuyXPayY(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {
	log.Println("=== calculateDiscountsBuyXPayY ...")
	if priceRuleVoucherPair.Rule.Action != ActionBuyXPayY {
		panic("CalculateDiscountsBuyXGetY called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	if calculationParameters.isCatalogCalculation == true {
		log.Println("catalog calculations can not handle actions of type CalculateDiscountsBuyXPayY")
		return orderDiscounts
	}

	//clone! we do not want to manipiulate cart/articleCollection item articleCollection
	var sortedPositions []Article
	for _, positionVoPtr := range calculationParameters.articleCollection.Articles {
		sortedPositions = append(sortedPositions, *positionVoPtr)
	}

	if priceRuleVoucherPair.Rule.WhichXYFree == XYMostExpensiveFree {
		sort.Sort(ByPriceDescending(sortedPositions))
	} else {
		sort.Sort(ByPriceAscending(sortedPositions))
	}

	//count matching first and articleCollection by price
	var totalMatchingQty float64
	for _, article := range sortedPositions {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, &article, calculationParameters, orderDiscounts)
		if ok {
			totalMatchingQty += article.Quantity
		}
	}

	var timesX = roundToStep(totalMatchingQty/float64(priceRuleVoucherPair.Rule.X), 0.01)
	timesXInt := int(math.Floor(timesX))
	freeQty := timesXInt * (priceRuleVoucherPair.Rule.X - int(priceRuleVoucherPair.Rule.Y))
	var productsAssignedFree int

	for _, article := range sortedPositions {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, &article, calculationParameters, orderDiscounts)

		orderDiscountsForPosition := orderDiscounts[article.ID]
		if !orderDiscounts[article.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition) {

			if freeQty > 0 {
				if productsAssignedFree < freeQty {
					//orderDiscountsForPosition := orderDiscounts[article.ID]
					//apply the discount here
					discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, article.ID)

					maxQty := int(article.Quantity)
					if maxQty > freeQty {
						maxQty = freeQty
					}

					for qty := 0; qty < int(maxQty); qty++ {
						//calculate the actual discount
						discountApplied.DiscountAmount += orderDiscounts[article.ID].CurrentItemPrice
						//discountApplied.DiscountAmount += article.Price
						//discountApplied.DiscountSingle += positionByPrice.Price // always zero as the discount is not for a single item
						discountApplied.Quantity = orderDiscounts[article.ID].Quantity
						productsAssignedFree++
						if productsAssignedFree >= freeQty {
							break
						}
					}
					orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, article.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, calculationParameters.roundTo)
					orderDiscounts[article.ID] = orderDiscountsForPosition
				}
			}
		}
	}
	return orderDiscounts
}

// ByPriceAscending implements sort.Interface for []Article based on
// the Price field.
type ByPriceAscending []Article

func (a ByPriceAscending) Len() int           { return len(a) }
func (a ByPriceAscending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriceAscending) Less(i, j int) bool { return a[i].Price < a[j].Price }

// ByPriceDescending implements sort.Interface for []Article based on
// the Price field.
type ByPriceDescending []Article

func (a ByPriceDescending) Len() int           { return len(a) }
func (a ByPriceDescending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriceDescending) Less(i, j int) bool { return a[i].Price > a[j].Price }
