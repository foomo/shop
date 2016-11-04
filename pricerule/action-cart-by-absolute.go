package pricerule

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/foomo/shop/order"
	"github.com/foomo/shop/utils"
)

// CalculateDiscountsCartByAbsolute -
func calculateDiscountsCartByAbsolute(order *order.Order, priceRuleVoucherPair *RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionCartByAbsolute {
		panic("CalculateDiscountsCartByAbsolute called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}
	//collect item values = price * qty for applicable items
	amounts := getAmountsOfApplicablePositions(priceRuleVoucherPair.Rule, order, productGroupIDsPerPosition, groupIDsForCustomer)

	// the tricky part - stolen code from Florian - distribute the amount proportional to the price
	distributedAmounts, err := Distribute(amounts, priceRuleVoucherPair.Rule.Amount)
	distribution := map[string]float64{}
	for i, distributedAmount := range distributedAmounts {
		distribution[order.GetPositions()[i].ItemID] = distributedAmount
	}

	if err != nil {
		panic(err)
	}

	for _, position := range order.Positions {

		// if we have the distributed amount
		if discountAmount, ok := distribution[position.ItemID]; ok {
			// and rule can still be applied
			if !orderDiscounts[position.ItemID].StopApplyingDiscounts && ok {
				if !orderDiscounts[position.ItemID].StopApplyingDiscounts {
					//apply the discount here
					discountApplied := &DiscountApplied{}
					discountApplied.PriceRuleID = priceRuleVoucherPair.Rule.ID
					discountApplied.MappingID = priceRuleVoucherPair.Rule.MappingID
					discountApplied.CalculationBasePrice = orderDiscounts[position.ItemID].CurrentItemPrice
					discountApplied.Price = orderDiscounts[position.ItemID].InitialItemPrice
					if priceRuleVoucherPair.Voucher != nil {
						discountApplied.VoucherID = priceRuleVoucherPair.Voucher.ID
						discountApplied.VoucherCode = priceRuleVoucherPair.Voucher.VoucherCode
					}

					//calculate the actual discount
					discountApplied.DiscountAmount = discountAmount
					discountApplied.DiscountSingle = discountAmount
					discountApplied.Quantity = orderDiscounts[position.ItemID].Qantity

					//pointer assignment WTF !!!
					orderDiscountsForPosition := orderDiscounts[position.ItemID]
					orderDiscountsForPosition.TotalDiscountAmount += discountApplied.DiscountAmount
					orderDiscountsForPosition.AppliedDiscounts = append(orderDiscountsForPosition.AppliedDiscounts, *discountApplied)
					orderDiscountsForPosition.CurrentItemPrice = utils.Round(discountApplied.CalculationBasePrice-discountAmount, 2)

					if priceRuleVoucherPair.Rule.Exclusive {
						orderDiscountsForPosition.StopApplyingDiscounts = true
					}
					orderDiscounts[position.ItemID] = orderDiscountsForPosition
				}
			}
		}
	}
	return orderDiscounts
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// Distribute - distribute the amounts proportionally to the total discount
func Distribute(amounts []float64, totalReduction float64) ([]float64, error) {
	amountsint64 := make([]int64, len(amounts))
	// Convert to Rappen and int64
	for i, amount := range amounts {
		amountsint64[i] = int64(amount * 100)
	}

	distribution := distributeI(amountsint64, int64(totalReduction*100))
	err := check(distribution, totalReduction)
	if err != nil {
		return distribution, err
	}
	return distribution, nil
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

// DistributeI calculates the reduction for each amount in amounts for totalReduction
func distributeI(amounts []int64, totalReduction int64) []float64 {
	var total int64
	for _, amount := range amounts {
		total += amount
	}

	amt1s := make([]int64, len(amounts))
	reductionInRappen := totalReduction // * 100

	for i := range amt1s {
		amt1s[i] = reductionInRappen * amounts[i] // * 100
	}
	reductions := distribute(total, amt1s)
	return adjustRoundingDifferences(float64(totalReduction*100), reductions)
}

func distribute(total int64, amt1s []int64) []int64 {
	var rappenDiff int64
	reductions := make([]int64, len(amt1s))

	for i := range reductions {
		reductions[i] = 0
	}

	for i := range amt1s {

		//fmt.Println(float64(amt1s[i])/float64(total), "Rounded:", Round(float64(amt1s[i])/float64(total)), "rappenDiff", rappenDiff)
		amt2 := int64(Round(float64(amt1s[i]) / float64(total))) // not perfect, but with the best results
		//amt2 := amt1s[i] / total // this would be the most correct version, but does not work
		remainder := int64(math.Mod(float64(amt1s[i]), float64(total)))
		condition := remainder*10 >= ((total * 10) / 2)
		amt3 := IteInt64(condition, amt2+1, amt2)
		reductions[i] = int64(Round(float64(amt3+rappenDiff)/5.0) * 5)
		//fmt.Println("Reduction", reductions[i])

		rappenDiff = amt3 - reductions[i] + rappenDiff
	}

	return reductions
}

// This does not work due to precision issues
func adjustRoundingDifferences(totalReduction float64, reductions []int64) []float64 {
	actualTotalReduction := 0.0
	for _, reduction := range reductions {
		actualTotalReduction += float64(reduction)
	}
	actualTotalReduction += 5
	reductionsDiff := totalReduction - actualTotalReduction
	// TODO: this part needs testing
	if int(math.Abs(float64(reductionsDiff))) == 5 {
		lastReduction := float64(reductions[len(reductions)-1])
		lastReductionAdjusted := lastReduction + reductionsDiff
		reductions[len(reductions)-1] = int64(lastReductionAdjusted)

		// TODO: this should be logged in the order history
		fmt.Println("Found potential rounding error (expected ", totalReduction, " but found ", actualTotalReduction, "), last partial reduction is set to ", lastReduction, " (from ", lastReductionAdjusted, ")")
	}

	reductionsF := make([]float64, len(reductions))
	for i, reduction := range reductions {
		reductionsF[i] = float64(reduction) / 100.0 // convert back from Rappen to Franken
	}
	return reductionsF
}

func check(reductions []float64, totalReduction float64) error {
	var sumReductions float64
	errString := ""
	for _, reduction := range reductions {
		sumReductions += reduction
	}
	if errString != "" {
		errString = "\n" + errString
	}

	sumReductions = roundToStep(sumReductions, 0.05) // this is necessary to get rid of tiny precision errors when added up floats
	if sumReductions != totalReduction {
		return errors.New("ERROR\n" + errString + "Total of distributed reduction is " + strconv.FormatFloat(sumReductions, 'f', 6, 64) + " but should be " + strconv.FormatFloat(totalReduction, 'f', 6, 64))
	}
	return nil

}

// RappenRound -
func RappenRound(value float64) float64 {
	tmp := 20 * value
	tmp = Round(tmp)
	return tmp / 20

}

// Round -
func Round(input float64) float64 {
	if input < 0 {
		return math.Ceil(input - 0.5)
	}
	return math.Floor(input + 0.5)
}

// IteInt64 -
// Ite => if then else. If true returns thenDo else elseDo
func IteInt64(condition bool, thenDo int64, elseDo int64) int64 {
	if condition {
		return thenDo
	}
	return elseDo
}

// get map of [positionID] => price*quantity for applicable positions
func getAmountsOfApplicablePositions(priceRule *PriceRule, order *order.Order, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string) []float64 {
	//collect item values = price * qty for applicable items
	amounts := []float64{}

	for _, position := range order.Positions {
		ok, _ := validatePriceRuleForPosition(*priceRule, order, position, productGroupIDsPerPosition, groupIDsForCustomer)
		if ok {
			amounts = append(amounts, position.Price*position.Quantity)
		}
	}
	return amounts
}
