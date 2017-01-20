package pricerule

import (
	"math"
	"sort"
)

// CalculateDiscountsBuyXGetY -
func calculateDiscountsBuyXGetY(itemCollVo *ItemCollection, priceRuleVoucherPair RuleVoucherPair, itemCollDiscounts OrderDiscounts, productGroupIDsPerItem map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionBuyXGetY {
		panic("CalculateDiscountsBuyXGetY called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	//count matching first and itemCollection by price
	var totalQty float64
	//clone! we do not want to manipiulate cart/itemCollection item itemCollection
	var sortedItems []Item
	for _, itemVoPtr := range itemCollVo.Items {
		sortedItems = append(sortedItems, *itemVoPtr)
	}
	if priceRuleVoucherPair.Rule.WhichXYFree == XYCheapestFree {
		sort.Sort(ByPriceAscending(sortedItems))
	} else {
		sort.Sort(ByPriceDescending(sortedItems))
	}

	for _, item := range itemCollVo.Items {
		ok, _ := validatePriceRuleForItem(*priceRuleVoucherPair.Rule, itemCollVo, item, productGroupIDsPerItem, groupIDsForCustomer)
		if ok {
			totalQty += item.Quantity
		}
	}

	var timesX = roundToStep(totalQty/float64(priceRuleVoucherPair.Rule.X), 0.01)
	timesXInt := int(math.Floor(timesX))
	freeQty := timesXInt * int(priceRuleVoucherPair.Rule.Y)

	if freeQty > 0 {
		var productsFree int
		for _, itemByPrice := range sortedItems {
			if productsFree < freeQty {
				itemCollDiscountsForItem := itemCollDiscounts[itemByPrice.ID]

				//apply the discount here
				discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, itemCollDiscounts, itemByPrice.ID)

				for qty := 0; qty < int(itemByPrice.Quantity); qty++ {
					//calculate the actual discount

					//calculate the actual discount
					discountApplied.DiscountAmount += itemByPrice.Price
					discountApplied.DiscountSingle += itemByPrice.Price
					discountApplied.Quantity = itemCollDiscounts[itemByPrice.ID].Quantity

					productsFree++
					if productsFree >= freeQty {
						break
					}
				}

				//add it to the ret obj
				//pointer assignment WTF !!!

				itemCollDiscountsForItem = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, itemByPrice.ID, itemCollDiscountsForItem, itemCollDiscounts, *priceRuleVoucherPair.Rule, roundTo)
				itemCollDiscounts[itemByPrice.ID] = itemCollDiscountsForItem

			}
			if productsFree >= freeQty {
				break
			}
		}

	}

	return itemCollDiscounts
}

// ByPriceAscending implements sort.Interface for []itemCollection.Item based on
// the Price field.
type ByPriceAscending []Item

func (a ByPriceAscending) Len() int           { return len(a) }
func (a ByPriceAscending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriceAscending) Less(i, j int) bool { return a[i].Price < a[j].Price }

// ByPriceDescending implements sort.Interface for []itemCollection.Item based on
// the Price field.
type ByPriceDescending []Item

func (a ByPriceDescending) Len() int           { return len(a) }
func (a ByPriceDescending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriceDescending) Less(i, j int) bool { return a[i].Price < a[j].Price }
