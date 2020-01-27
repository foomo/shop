package pricerule

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type cumulationTestHelper struct {
	CustomerGroupRegular  string
	CustomerGroupEmployee string
	Sku1                  string
	Sku2                  string
	Sku3                  string
	GroupIDSingleSku1     string
	GroupIDSingleSku2     string
	GroupIDTwoSkus        string
	GroupIDThreeSkus      string
}

func newTesthelper() cumulationTestHelper {
	return cumulationTestHelper{
		CustomerGroupRegular:  "customer-regular",
		CustomerGroupEmployee: "customer-employee",
		Sku1:                  "sku1",
		Sku2:                  "sku2",
		GroupIDSingleSku1:     "group-with-sku1",
		GroupIDSingleSku2:     "group-with-sku2",
		GroupIDTwoSkus:        "group-with-two-skus",
		GroupIDThreeSkus:      "group-with-three-skus",
	}
}

func (helper cumulationTestHelper) cleanupTestData(t *testing.T) {
	assert.NoError(t, RemoveAllGroups())
	assert.NoError(t, RemoveAllPriceRules())
	assert.NoError(t, RemoveAllVouchers())
	ClearCache() // reset cache for catalogue calculations
}
func (helper cumulationTestHelper) cleanupAndRecreateTestData(t *testing.T) {
	helper.cleanupTestData(t)
	productsInGroups := make(map[string][]string)
	productsInGroups[helper.GroupIDSingleSku1] = []string{helper.Sku1}
	productsInGroups[helper.GroupIDSingleSku2] = []string{helper.Sku2}
	productsInGroups[helper.GroupIDTwoSkus] = []string{helper.Sku1, helper.Sku2}
	productsInGroups[helper.GroupIDThreeSkus] = []string{helper.Sku1, helper.Sku2, helper.Sku3}

	helper.createMockCustomerGroups(t, []string{helper.CustomerGroupRegular, helper.CustomerGroupEmployee})
	helper.createMockProductGroups(t, productsInGroups)
}

func (helper cumulationTestHelper) createMockProductGroups(t *testing.T, productGroups map[string][]string) {
	for groupID, items := range productGroups {
		group := new(Group)
		group.Type = ProductGroup
		group.ID = groupID
		group.Name = groupID
		assert.NoError(t, group.AddGroupItemIDsAndPersist(items), "Could not create product groups")
	}
}

func (helper cumulationTestHelper) createMockCustomerGroups(t *testing.T, customerGroups []string) {
	for _, groupID := range customerGroups {
		group := new(Group)
		group.Type = CustomerGroup
		group.ID = groupID
		group.Name = groupID
		assert.NoError(t, group.AddGroupItemIDsAndPersist([]string{groupID}), "Could not create customer groups")
	}
}

// createVouchers creates n voucher codes for given Pricerule
func (helper cumulationTestHelper) createMockVouchers(t *testing.T, priceRule *PriceRule, n int) []string {
	vouchers := make([]string, n)
	for i, _ := range vouchers {
		voucherCode := "voucher-" + priceRule.ID + "-" + strconv.Itoa(i)
		voucher := NewVoucher(voucherCode, voucherCode, priceRule, "")
		assert.NoError(t, voucher.Upsert())
		vouchers[i] = voucherCode
	}
	return vouchers
}

func (helper cumulationTestHelper) getMockArticleCollection() *ArticleCollection {
	return &ArticleCollection{
		Articles: []*Article{
			{
				ID:                         helper.Sku1,
				Price:                      10.0,
				Quantity:                   1,
				AllowCrossPriceCalculation: true,
			},
			{
				ID:                         helper.Sku2,
				Price:                      20.0,
				Quantity:                   1,
				AllowCrossPriceCalculation: true,
			},
		},
		CustomerType: helper.CustomerGroupRegular,
	}
}
func (helper cumulationTestHelper) setMockEmployeeDiscount10Percent(t *testing.T, name string, includedProductGroupIDS []string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRule-EmployeeDiscount-" + name)
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupEmployee}
	assert.NoError(t, priceRule.Upsert())

	return
}
func (helper cumulationTestHelper) setMockPriceRuleAndVoucherXPercent(t *testing.T, name string, amount float64, includedProductGroupIDS []string, excludeAlreadyDiscountedItems bool, excludeEmployees bool) string {

	// Create pricerule
	priceRule := NewPriceRule("PriceRule-" + name)
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	if excludeEmployees {
		priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupRegular}
	}
	priceRule.ExcludeAlreadyDiscountedItemsForVoucher = excludeAlreadyDiscountedItems
	assert.Nil(t, priceRule.Upsert())

	// Create voucher
	return helper.createMockVouchers(t, priceRule, 1)[0]
}

func (helper cumulationTestHelper) setMockPriceRuleCrossPrice(t *testing.T, name string, amount float64, isPercent bool, includedProductGroupIDS []string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRulePromotionProduct-" + name)
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	if isPercent {
		priceRule.Action = ActionItemByPercent
	}
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	assert.NoError(t, priceRule.Upsert())
}
func (helper cumulationTestHelper) setMockPriceRuleBuy3Pay2(t *testing.T, name string, includedProductGroupIDS []string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRuleBuy3Pay2-" + name)
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionBuyXPayY
	priceRule.X = 3
	priceRule.Y = 2
	priceRule.WhichXYFree = XYCheapestFree
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	assert.NoError(t, priceRule.Upsert())
}

func (helper cumulationTestHelper) setMockPriceRuleAndVoucherAbsoluteCHF20(t *testing.T, cumulate bool) (string, string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRule-" + "CartAbsoluteCHF20")
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByAbsolute
	priceRule.Amount = 20.0
	priceRule.IncludedProductGroupIDS = []string{helper.GroupIDTwoSkus}
	priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupRegular}
	priceRule.ExcludeAlreadyDiscountedItemsForVoucher = false
	priceRule.CumulateWithOtherVouchers = cumulate
	assert.NoError(t, priceRule.Upsert())

	// Create vouchers
	vouchers := helper.createMockVouchers(t, priceRule, 2)
	return vouchers[0], vouchers[1]
}
func (helper cumulationTestHelper) setMockPriceRuleAndVoucher10Percent(t *testing.T, cumulate bool) string {

	// Create pricerule
	priceRule := NewPriceRule("PriceRule-" + "Cart10Percent")
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10.0
	priceRule.IncludedProductGroupIDS = []string{helper.GroupIDTwoSkus}
	priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupRegular}
	priceRule.ExcludeAlreadyDiscountedItemsForVoucher = false
	priceRule.CumulateWithOtherVouchers = cumulate
	assert.NoError(t, priceRule.Upsert())

	// Create voucher
	return helper.createMockVouchers(t, priceRule, 1)[0]
}

func (helper cumulationTestHelper) setMockPriceRuleQtyThreshold(t *testing.T, name string, amount float64, threshold float64, includedProductGroupIDS []string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRulePromotionQtyThreshold-" + name)
	priceRule.Type = TypePromotionOrder // TypePromotionProduct would also work here
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.QtyThreshold = threshold
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	assert.NoError(t, priceRule.Upsert())
}
func (helper cumulationTestHelper) setMockPriceRuleThresholdAmount(t *testing.T, name string, amount float64, thresholdAmount float64, includedProductGroupIDS []string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRulePromotionProductThresholdAmount-" + name)
	priceRule.Type = TypePromotionOrder // TypePromotionProduct would also work here
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = amount
	priceRule.MinOrderAmount = thresholdAmount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	assert.NoError(t, priceRule.Upsert())
}

func (helper cumulationTestHelper) setMockBonusVoucherPriceRule(t *testing.T) (string, string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRule-" + "Bonus10CHF")
	priceRule.Description = priceRule.Name
	priceRule.Type = TypeBonusVoucher
	priceRule.Action = ActionCartByAbsolute
	priceRule.Amount = 10.0

	priceRule.CalculateDiscountedOrderAmount = true
	priceRule.Priority = 999
	assert.NoError(t, priceRule.Upsert())

	// Create vouchers
	vouchers := helper.createMockVouchers(t, priceRule, 2)
	return vouchers[0], vouchers[1]
}

func (helper cumulationTestHelper) accumulateDiscountsOfItems(discounts OrderDiscounts) float64 {
	sum := 0.0
	for _, d := range discounts {
		sum += d.TotalDiscountAmountApplicable
	}
	return sum
}
