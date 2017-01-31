package pricerule

import (
	"fmt"
	"log"
	"math"
)

// CalculateDiscountsCartByAbsolute -
func calculateDiscountsCartByAbsolute(articleCollection *ArticleCollection, priceRuleVoucherPair RuleVoucherPair, orderDiscounts OrderDiscounts, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, roundTo float64, isCatalogCalculation bool) OrderDiscounts {
	if priceRuleVoucherPair.Rule.Action != ActionCartByAbsolute {
		panic("CalculateDiscountsCartByAbsolute called with pricerule of action " + priceRuleVoucherPair.Rule.Action)
	}

	//collect item values = price * qty for applicable items
	amountsMap := getAmountsOfApplicablePositions(priceRuleVoucherPair.Rule, articleCollection, productGroupIDsPerPosition, groupIDsForCustomer, isCatalogCalculation)
	amounts := getMapValues(amountsMap)
	//spew.Dump(amounts)
	// the tricky part - stolen code from Florian - distribute the amount proportional to the price
	distributedAmounts, err := Distribute(amounts, priceRuleVoucherPair.Rule.Amount)
	distribution := map[string]float64{}

	i := 0
	for itemID, _ := range amountsMap {
		distribution[itemID] = distributedAmounts[i]
		i++
	}
	fmt.Println("===> promo distribution")
	fmt.Println(distribution)
	if err != nil {
		panic(err)
	}

	for _, article := range articleCollection.Articles {
		// if we have the distributed amount
		if discountAmount, ok := distribution[article.ID]; ok {
			// and rule can still be applied
			orderDiscountsForPosition := orderDiscounts[article.ID]
			if !orderDiscounts[article.ID].StopApplyingDiscounts && ok && !previouslyAppliedExclusionInPlace(priceRuleVoucherPair.Rule, orderDiscountsForPosition) {
				if !orderDiscounts[article.ID].StopApplyingDiscounts {
					//apply the discount here
					discountApplied := getInitializedDiscountApplied(priceRuleVoucherPair, orderDiscounts, article.ID)

					//calculate the actual discount
					discountApplied.DiscountAmount = discountAmount
					discountApplied.DiscountSingle = discountAmount / orderDiscounts[article.ID].Quantity
					discountApplied.Quantity = orderDiscounts[article.ID].Quantity

					//pointer assignment WTF !!!
					orderDiscountsForPosition := orderDiscounts[article.ID]
					orderDiscountsForPosition = calculateCurrentPriceAndApplicableDiscountsEnforceRules(*discountApplied, article.ID, orderDiscountsForPosition, orderDiscounts, *priceRuleVoucherPair.Rule, roundTo)

					orderDiscounts[article.ID] = orderDiscountsForPosition
				}
			}
		}
	}
	return orderDiscounts
}

// Get values from map
func getMapValues(mapVal map[string]float64) []float64 {
	ret := []float64{}
	for _, val := range mapVal {
		ret = append(ret, val)
	}
	return ret
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

		// TODO: this should be logged in the articleCollection history
		fmt.Println("Found potential rounding error (expected ", totalReduction, " but found ", actualTotalReduction, "), last partial reduction is set to ", lastReduction, " (from ", lastReductionAdjusted, ")")
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
		log.Println("WARNING: Total of distributed reduction has to be corrected by ", diff)
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
func getAmountsOfApplicablePositions(priceRule *PriceRule, articleCollection *ArticleCollection, productGroupIDsPerPosition map[string][]string, groupIDsForCustomer []string, isCatalogCalculation bool) map[string]float64 {
	//collect item values = price * qty for applicable items
	//amounts := []float64{}
	amountsMap := make(map[string]float64)

	for _, article := range articleCollection.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRule, articleCollection, article, productGroupIDsPerPosition, groupIDsForCustomer, isCatalogCalculation)
		if ok {
			//amounts = append(amounts, article.Price*article.Quantity)
			amountsMap[article.ID] = article.Price * article.Quantity
		}
	}
	//return amounts, amountsMap
	return amountsMap
}
