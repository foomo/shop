package pricerule

import (
	"math"
	"sort"
)

// CalculateDiscountsBuyXGetY -
func calculateDiscountsBuyXGetY(orderVo *ArticleCollection, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionBuyXGetY {
		panic("CalculateDiscountsBuyXGetY called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	//count matching first and articleCollection by price
	var totalQty float64
	//clone! we do not want to manipiulate cart/articleCollection item articleCollection
	var sortedPositions []Article
	for _, positionVoPtr := range orderVo.Articles {
		sortedPositions = append(sortedPositions, *positionVoPtr)
	}
	if priceRuleVoucherPair.Rule.WhichXYFree == XYCheapestFree {
		sort.Sort(ByPriceAscending(sortedPositions))
	} else {
		sort.Sort(ByPriceDescending(sortedPositions))
	}

	for _, article := range orderVo.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, orderVo, article, productGroupIDsPerPosition, groupIDsForCustomer)
		if ok {
			totalQty += article.Quantity
		}
	}

	var timesX = roundToStep(totalQty/float64(priceRuleVoucherPair.Rule.X), 0.01)
	timesXInt := int(math.Floor(timesX))
	freeQty := timesXInt * int(priceRuleVoucherPair.Rule.Y)

	if freeQty > 0 {
		var productsFree int
		for _, positionByPrice := range sortedPositions {
			if productsFree < freeQty {
				orderDiscountsForPosition := orderDiscounts[positionByPrice.ID]

				//apply the discount here
				discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, positionByPrice.ID)

				for qty := 0; qty < int(positionByPrice.Quantity); qty++ {
					//calculate the actual discount

					//calculate the actual discount
					discountApplied.DiscountAmount += positionByPrice.Price
					discountApplied.DiscountSingle += positionByPrice.Price
					discountApplied.Quantity = orderDiscounts[positionByPrice.ID].Quantity

					productsFree++
					if productsFree >= freeQty {
						break
					}
				}

				//add it to the ret obj
				//pointer assignment WTF !!!

				orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, positionByPrice.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, roundTo)
				orderDiscounts[positionByPrice.ID] = orderDiscountsForPosition

			}
			if productsFree >= freeQty {
				break
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
func (a ByPriceDescending) Less(i, j int) bool { return a[i].Price < a[j].Price }
