package pricerule

import (
	"fmt"
	"log"
	"math"
)

// CalculateDiscountsCartByAbsolute -
func calculateDiscountsCartByAbsolute(priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, calculationParameters *CalculationParameters) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionCartByAbsolute {
		panic("CalculateDiscountsCartByAbsolute called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	if calculationParameters.isCatalogCalculation == true {
		if Verbose {
			log.Println("catalog calculations can not handle actions of type CalculateDiscountsCartByAbsolute")
		}
		return orderDiscounts
	}

	//collect item values = price * qty for applicable items
	amountsMap := getAmountsOfApplicablePositions(priceRuleVoucherPair.Rule, calculationParameters, orderDiscounts)
	itemIDs, amounts := getMapValues(amountsMap)

	// the tricky part - stolen code from Florian - distribute the amount proportional to the price
	distributedAmounts, err := Distribute(amounts, priceRuleVoucherPair.Rule.Amount)
	distribution := map[string]float64{}

	for i, itemID := range itemIDs {
		distribution[itemID] = distributedAmounts[i]
	}
	if Verbose {
		fmt.Println("===> promo distribution")
		fmt.Println(distribution)
	}
	if err != nil {
		panic(err)
	}

	for _, article := range calculationParameters.articleCollection.Articles {
		// if we have the distributed amount
		if discountAmount, ok := distribution[article.ID]; ok {
			// and rule can still be applied
			orderDiscountsForPosition := orderDiscounts[article.ID]
			if !orderDiscounts[article.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition, calculationParameters) {
				if !orderDiscounts[article.ID].StopApplyingDiscounts {
					//apply the discount here
					discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, article.ID)

					//calculate the actual discount
					discountApplied.DiscountAmount = discountAmount
					discountApplied.DiscountSingle = discountAmount / orderDiscounts[article.ID].Quantity
					discountApplied.Quantity = orderDiscounts[article.ID].Quantity

					//pointer assignment WTF !!!
					orderDiscountsForPosition := orderDiscounts[article.ID]
					orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, article.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, calculationParameters.roundTo)
					orderDiscounts[article.ID] = orderDiscountsForPosition
				}
			}
		}
	}
	return orderDiscounts
}

// Get values from map
func getMapValues(mapVal map[string]float64) ([]string, []float64) {
	vals := []float64{}
	keys := []string{}

	for key, val := range mapVal {
		keys = append(keys, key)
		vals = append(vals, val)
	}

	return keys, vals
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// Distribute - distribute the amounts proportionally to the total discount
func Distribute(amounts []float64, totalReduction float64) ([]float64, error) {
	if len(amounts) == 0 {
		return []float64{}, nil
	}
	amountsint64 := make([]int64, len(amounts))
	// Convert to Rappen and int64
	for i, amount := range amounts {
		amountsint64[i] = int64(amount * 100)
	}

	distribution := distributeI(amountsint64, int64(totalReduction*100))
	diff := check(distribution, totalReduction)
	if diff > 0 {
		distribution[len(distribution)-1] = distribution[len(distribution)-1] - diff
	} else {
		distribution[len(distribution)-1] = distribution[len(distribution)-1] + diff

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

		amt2 := int64(Round(float64(amt1s[i]) / float64(total))) // not perfect, but with the best results
		remainder := int64(math.Mod(float64(amt1s[i]), float64(total)))
		condition := remainder*10 >= ((total * 10) / 2)
		amt3 := IteInt64(condition, amt2+1, amt2)
		reductions[i] = int64(Round(float64(amt3+rappenDiff)/5.0) * 5)

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

		if Verbose {
			fmt.Println("Found potential rounding error (expected ", totalReduction, " but found ", actualTotalReduction, "), last partial reduction is set to ", lastReduction, " (from ", lastReductionAdjusted, ")")
		}
	}

	reductionsF := make([]float64, len(reductions))
	for i, reduction := range reductions {
		reductionsF[i] = float64(reduction) / 100.0 // convert back from Rappen to Franken
	}
	return reductionsF
}

func check(reductions []float64, totalReduction float64) float64 {
	var sumReductions float64

	for _, reduction := range reductions {
		sumReductions += reduction
	}

	sumReductions = roundToStep(sumReductions, 0.05) // this is necessary to get rid of tiny precision errors when added up floats
	if sumReductions != totalReduction {
		diff := roundToStep(sumReductions-totalReduction, 0.05)
		if Verbose {
			log.Println("WARNING: Total of distributed reduction has to be corrected by ", diff)
		}
		return diff
	}
	return 0

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
func getAmountsOfApplicablePositions(priceRule *PriceRule, calculationParameters *CalculationParameters, orderDiscounts OrderDiscounts) map[string]float64 {
	//collect item values = price * qty for applicable items
	amountsMap := make(map[string]float64)
	if len(priceRule.ItemSets) == 0 {
		for _, article := range calculationParameters.articleCollection.Articles {

			ok, _ := validatePriceRuleForPosition(*priceRule, article, calculationParameters, orderDiscounts)
			if ok {
				//amounts = append(amounts, article.Price*article.Quantity)
				amountsMap[article.ID] = article.Price * article.Quantity
			}
		}
	} else {
		// if ActionCartByAbsolute is used to implement ActionItemSetByAbolute make sure we distribute to the right items
		for _, article := range calculationParameters.articleCollection.Articles {

			if isItemInItemSet(priceRule.ItemSets, article.ID) {
				//amounts = append(amounts, article.Price*article.Quantity)
				amountsMap[article.ID] = article.Price * article.Quantity
			}
		}

	}
	//return amounts, amountsMap
	return amountsMap
}

func isItemInItemSet(itemSets [][]string, itemID string) (ret bool) {
	ret = false
	for _, items := range itemSets {
		if contains(itemID, items) {
			return true
		}
	}

	return ret
}
