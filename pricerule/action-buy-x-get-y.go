package pricerule

import (
	"log"
	"math"
)

// CalculateDiscountsBuyXGetY -
func calculateDiscountsBuyXGetY(articleCollection *ArticleCollection, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64, isCatalogCalculation bool) OrderDiscounts {
	log.Println("=== calculateDiscountsBuyXGetY ...")
	if priceRuleVoucherPair.Rule.Action != ActionBuyXGetY {
		panic("CalculateDiscountsBuyXGetY called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	if isCatalogCalculation == true {
		panic("catalog calculations can not handle actions of type CalculateDiscountsBuyXGetY")
	}

	for _, article := range articleCollection.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, articleCollection, article, productGroupIDsPerPosition, groupIDsForCustomer, isCatalogCalculation)

		orderDiscountsForPosition := orderDiscounts[article.ID]
		if !orderDiscounts[article.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition) {

			totalQty := article.Quantity
			var timesX = roundToStep(totalQty/float64(priceRuleVoucherPair.Rule.X), 0.01)
			timesXInt := int(math.Floor(timesX))
			freeQty := timesXInt * int(priceRuleVoucherPair.Rule.Y)
			//	fmt.Println("freeQty", freeQty)
			//	if freeQty > 0 {
			var productsFree int

			if productsFree < freeQty {
				//orderDiscountsForPosition := orderDiscounts[article.ID]

				//apply the discount here
				discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, article.ID)

				for qty := 0; qty < int(article.Quantity/float64(priceRuleVoucherPair.Rule.Y)); qty++ {
					//calculate the actual discount
					discountApplied.DiscountAmount += orderDiscounts[article.ID].CurrentItemPrice
					//discountApplied.DiscountAmount += article.Price
					//discountApplied.DiscountSingle += positionByPrice.Price // always zero as the discount is not for a single item
					discountApplied.Quantity = orderDiscounts[article.ID].Quantity

					productsFree++
					if productsFree >= freeQty {
						break
					}
				}

				orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, article.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, roundTo)

				orderDiscounts[article.ID] = orderDiscountsForPosition

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
