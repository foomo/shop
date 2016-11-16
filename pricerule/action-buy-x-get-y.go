package pricerule

import (
	"math"
	"sort"

	"github.com/foomo/shop/order"
)

// CalculateDiscountsBuyXGetY -
func calculateDiscountsBuyXGetY(orderVo *order.Order, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionBuyXGetY {
		panic("CalculateDiscountsBuyXGetY called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	//count matching first and order by price
	var totalQty float64
	//clone! we do not want to manipiulate cart/order item order
	var sortedPositions []order.Position
	for _, positionVoPtr := range orderVo.Positions {
		sortedPositions = append(sortedPositions, *positionVoPtr)
	}
	if priceRuleVoucherPair.Rule.WhichXYFree == XYCheapestFree {
		sort.Sort(ByPriceAscending(sortedPositions))
	} else {
		sort.Sort(ByPriceDescending(sortedPositions))
	}

	for _, position := range orderVo.Positions {
		ok, _ := validatePriceRuleForPosition(*priceRuleVoucherPair.Rule, orderVo, position, productGroupIDsPerPosition, groupIDsForCustomer)
		if ok {
			totalQty += position.Quantity
		}
	}

	var timesX = roundToStep(totalQty/float64(priceRuleVoucherPair.Rule.X), 0.01)
	timesXInt := int(math.Floor(timesX))
	freeQty := timesXInt * int(priceRuleVoucherPair.Rule.Y)

	if freeQty > 0 {
		var productsFree int
		for _, positionByPrice := range sortedPositions {

			if productsFree < freeQty {

				//apply the discount here
				discountApplied := &DiscountApplied{}
				discountApplied.PriceRuleID = priceRuleVoucherPair.Rule.ID
				discountApplied.MappingID = priceRuleVoucherPair.Rule.MappingID
				discountApplied.CalculationBasePrice = orderDiscounts[positionByPrice.ItemID].CurrentItemPrice
				discountApplied.Price = orderDiscounts[positionByPrice.ItemID].InitialItemPrice

				for qty := 0; qty < int(positionByPrice.Quantity); qty++ {
					//calculate the actual discount
					discountApplied.DiscountAmount += positionByPrice.Price
					productsFree++
					if productsFree >= freeQty {
						break
					}
				}

				if priceRuleVoucherPair.Voucher != nil {
					discountApplied.VoucherID = priceRuleVoucherPair.Voucher.ID
					discountApplied.VoucherCode = priceRuleVoucherPair.Voucher.VoucherCode
				}

				//add it to the ret obj
				//pointer assignment WTF !!!
				orderDiscountsForPosition := orderDiscounts[positionByPrice.ItemID]
				orderDiscountsForPosition.TotalDiscountAmount += discountApplied.DiscountAmount
				orderDiscountsForPosition.AppliedDiscounts = append(orderDiscountsForPosition.AppliedDiscounts, *discountApplied)
				if priceRuleVoucherPair.Rule.Exclusive {
					orderDiscountsForPosition.StopApplyingDiscounts = true
				}
				orderDiscounts[positionByPrice.ItemID] = orderDiscountsForPosition

			}
			if productsFree >= freeQty {
				break
			}
		}

	}

	return orderDiscounts
}

// ByPriceAscending implements sort.Interface for []order.Position based on
// the Price field.
type ByPriceAscending []order.Position

func (a ByPriceAscending) Len() int           { return len(a) }
func (a ByPriceAscending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriceAscending) Less(i, j int) bool { return a[i].Price < a[j].Price }

// ByPriceDescending implements sort.Interface for []order.Position based on
// the Price field.
type ByPriceDescending []order.Position

func (a ByPriceDescending) Len() int           { return len(a) }
func (a ByPriceDescending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriceDescending) Less(i, j int) bool { return a[i].Price < a[j].Price }
