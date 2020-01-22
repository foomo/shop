package pricerule

import (
	"strconv"
	"testing"

	"github.com/foomo/shop/utils"
	"github.com/stretchr/testify/assert"
)

type cumulationTestHelper struct {
	CustomerGroupRegular  string
	CustomerGroupEmployee string
	Sku1                  string
	Sku2                  string
	GroupIDSingleSku1     string
	GroupIDSingleSku2     string
	GroupIDTwoSkus        string
}

func newTesthelper() cumulationTestHelper {
	return cumulationTestHelper{
		CustomerGroupRegular:  "customer-regular",
		CustomerGroupEmployee: "customer-employee",
		Sku1:                  "sku1",
		Sku2:                  "sku2",
		GroupIDSingleSku1:     "group-with-sku1",
		GroupIDSingleSku2:     "group-with-sku2",
		GroupIDTwoSkus:        "group-with-both-skus",
	}
}

func (helper cumulationTestHelper) cleanupTestData(t *testing.T) {
	assert.NoError(t, RemoveAllGroups())
	assert.NoError(t, RemoveAllPriceRules())
	assert.NoError(t, RemoveAllVouchers())
}
func (helper cumulationTestHelper) cleanupAndRecreateTestData(t *testing.T) {
	helper.cleanupTestData(t)
	productsInGroups := make(map[string][]string)
	productsInGroups[helper.GroupIDSingleSku1] = []string{helper.Sku1}
	productsInGroups[helper.GroupIDSingleSku2] = []string{helper.Sku2}
	productsInGroups[helper.GroupIDTwoSkus] = []string{helper.Sku1, helper.Sku2}

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

func (helper cumulationTestHelper) setMockPriceRuleCrossPrice(t *testing.T, name string, amount float64, includedProductGroupIDS []string) {

	// Create pricerule
	priceRule := NewPriceRule("PriceRulePromotionProduct-" + name)
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = amount
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

func TestCumulationTwoVouchers_OnePerSku(t *testing.T) {

	// Cart with 2 Items
	// Expected result: one of the vouchers is applied to each item

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDSingleSku1}, false, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku2", 10.0, []string{helper.GroupIDSingleSku2}, false, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 10% on 30 CHF
	assert.Equal(t, 3.0, summary.TotalDiscountApplicable)
}
func TestCumulationTwoVouchers_BothForSameSku(t *testing.T) {

	// Cart with 2 Items
	// Both voucher are only valid for one of the items
	// Expected result: One voucher is being applied, the other one is dismissed

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDSingleSku1}, false, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku2", 10.0, []string{helper.GroupIDSingleSku1}, false, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// only one voucher should be applied => 1 CHF
	assert.Equal(t, 1.0, summary.TotalDiscountApplicable)

}
func TestCumulationTwoVouchers_BothForBothSkus(t *testing.T) {

	// Cart with 2 Items
	// Both voucher are  valid for both of the items
	// Expected result: The better voucher is applied to both items

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-1", 5.0, []string{helper.GroupIDTwoSkus}, false, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-2", 10.0, []string{helper.GroupIDTwoSkus}, false, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 10% on 30 CHF
	assert.Equal(t, 3.0, summary.TotalDiscountApplicable)

}
func TestCumulationTwoVouchers_BothForSameSku_AdditonalCrossPrice(t *testing.T) {

	// Cart with 2 Items
	// Both voucher are only valid for one of the items
	// Expected result: - both items discounted by product promo
	//					- One voucher is being applied, the other one is dismissed

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0, []string{helper.GroupIDTwoSkus})
	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 5.0, []string{helper.GroupIDSingleSku1}, false, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku2", 10.0, []string{helper.GroupIDSingleSku1}, false, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// helper.Sku1: 5 + 0.5, helper.Sku2: 5 => 10.5
	assert.Equal(t, 10.5, summary.TotalDiscountApplicable)

}
func TestCumulationProductPromo(t *testing.T) {

	// Cart with 2 Items
	// Expected result: - only better cross price should be applied

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 2.0, []string{helper.GroupIDTwoSkus})
	helper.setMockPriceRuleCrossPrice(t, "crossprice2", 5.0, []string{helper.GroupIDTwoSkus})

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, []string{}, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 5+5 (5 on each product) = 10 CHF
	assert.Equal(t, 10.0, summary.TotalDiscountApplicable)

}

func TestCumulationForExcludeVoucherOnCrossPriceWebhop(t *testing.T) {
	// Cart with 2 Items
	// Expected result: voucher is only applied to sku2. helper.Sku1 is skipped due to existing webshop cross-price

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0, []string{helper.GroupIDSingleSku1})
	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-1", 10.0, []string{helper.GroupIDTwoSkus}, true, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// helper.Sku1: 5, helper.Sku2: 2
	assert.Equal(t, 7.0, summary.TotalDiscountApplicable)

}
func TestCumulationForExcludeVoucherOnCrossPriceSAP(t *testing.T) {

	// Cart with 2 Items
	// Expected result: voucher is only applied to sku2. helper.Sku1 is skipped due to existing SAP cross-price
	// (indicated by AllowCrossPriceCalculation == false)

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()
	for i, article := range articleCollection.Articles {
		if i == 0 {
			article.AllowCrossPriceCalculation = false // this is false if there already is a SAP crossprice
		}
	}

	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-1", 10.0, []string{helper.GroupIDTwoSkus}, true, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// helper.Sku2: 2 CHF
	assert.Equal(t, 2.0, summary.TotalDiscountApplicable)

}

func TestCumulationTwoVouchers_BothForBothSkus_BothApplied(t *testing.T) {

	// Cart with 2 Items (CHF 20 and CHF 10)
	// Both voucher are  valid for both of the items
	// Expected result: Both voucher are applied because on voucher has set CumulateWithOtherVouchers = true

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1, _ := helper.setMockPriceRuleAndVoucherAbsoluteCHF20(t, true)
	voucherCode2 := helper.setMockPriceRuleAndVoucher10Percent(t, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 20 CHF (absolute) + 3 (10%) = 23
	assert.Equal(t, 23.0, summary.TotalDiscountApplicable)
}
func TestCumulationTwoVouchers_BothForBothSkus_SamePromo_BothApplied(t *testing.T) {

	// Cart with 2 Items (CHF 20 and CHF 10)
	// Both voucher are  valid for both of the items
	// Vouchers are of same promo
	// CumulateWithOtherVouchers = true
	// Expected result: Both voucher are applied because CumulateWithOtherVouchers = true

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1, voucherCode2 := helper.setMockPriceRuleAndVoucherAbsoluteCHF20(t, true)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 20 CHF (absolute) + 10 (20 CHF but only 10 applicable) = 30
	assert.Equal(t, 30.0, summary.TotalDiscountApplicable)

}
func TestCumulationTwoVouchers_BothForBothSkus_SamePromoOnlyOneApplied(t *testing.T) {

	// Cart with 2 Items (CHF 20 and CHF 10)
	// Both voucher are  valid for both of the items
	// Vouchers are of same promo
	// CumulateWithOtherVouchers = false
	// Expected result: Only one voucher is applied because CumulateWithOtherVouchers = true

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1, voucherCode2 := helper.setMockPriceRuleAndVoucherAbsoluteCHF20(t, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 20 CHF (absolute)  0 = 20
	assert.Equal(t, 20.0, summary.TotalDiscountApplicable)

}
func TestCumulationBonusVoucher(t *testing.T) {

	// Cart with 2 Items (CHF 20 and CHF 10)
	// Both bonus vouchers should be applied
	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupTestData(t)

	articleCollection := helper.getMockArticleCollection()

	voucherCode1, voucherCode2 := helper.setMockBonusVoucherPriceRule(t)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 10 +10 = 20 CHF
	assert.Equal(t, 20.0, summary.TotalDiscountApplicable)

}

func TestCumulationExcludeEmployeesFromVoucher(t *testing.T) {

	// Cart with 2 Items
	// Expected result: only regular customer gets discount

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDTwoSkus}, false, true)

	// Get Discounts for regular customer
	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 10% on 30 CHF
	assert.Equal(t, 3.0, summary.TotalDiscountApplicable)

	// Change customer type to employee => no discount
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	discounts, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// No discount for employee
	assert.Equal(t, 0.0, summary.TotalDiscountApplicable)

}
func TestCumulationEmployeeDiscount(t *testing.T) {

	// Cart with 2 Items
	// Expected result: only employee customer gets discount

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockEmployeeDiscount10Percent(t, "employee-discount", []string{helper.GroupIDTwoSkus})

	// no employee discuonts for regular customer
	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	assert.Equal(t, 0.0, summary.TotalDiscountApplicable)

	// Change customer type to eomplyee => discount applied
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	discounts, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// No discount for non-regular customer
	assert.Equal(t, 3.0, summary.TotalDiscountApplicable)

}
func TestCumulationCrossPriceAndEmployeeDiscount(t *testing.T) {

	// Cart with 2 Items
	// Expected result: regular customer gets crossprice, employee gets additional employee discount

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, []string{helper.GroupIDTwoSkus})
	helper.setMockEmployeeDiscount10Percent(t, "employee-discount", []string{helper.GroupIDTwoSkus})

	// no employee discounts for regular customer
	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 5+5 CHF Crossprices, no employee discount
	assert.Equal(t, 10.0, summary.TotalDiscountApplicable)

	// Change customer type to eomplyee => discount applied
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	discounts, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 5+5 CHF Crossprices + 2 CHF employee discount
	assert.Equal(t, 12.0, summary.TotalDiscountApplicable)

}
func TestCumulationEmployeeDiscountAndVoucherIncluceDiscountedItems(t *testing.T) {

	// - Cart with 2 Items
	// - Cross-price on Sku1
	// - Employee discount on both items
	// - 10% Voucher
	// Expected result: Employee discount is granted regardless of flag ExcludeAlreadyDiscountedItemsForVoucher

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, []string{helper.GroupIDSingleSku1})
	helper.setMockEmployeeDiscount10Percent(t, "employee-discount", []string{helper.GroupIDTwoSkus})
	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDTwoSkus}, false, false)
	// no employee discounts for regular customer
	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)
	// 5 CHF crossprice for Sku1 + 2.5 CHF voucher discount
	assert.Equal(t, 7.5, summary.TotalDiscountApplicable)

	// Change customer type to employee => additional employee discount
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	discounts, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)
	// 5 CHF crossprice for Sku1 + 2.5 CHF employee discount + 2.25 CHF voucher discount
	assert.Equal(t, 9.75, summary.TotalDiscountApplicable)

}
func TestCumulationEmployeeDiscountAndVoucherExcludeDiscountedItems(t *testing.T) {

	// - Cart with 2 Items
	// - Cross-price on Sku1
	// - Employee discount on both items
	// - 10% Voucher
	// Expected result: Employee discount is granted regardless of flag ExcludeAlreadyDiscountedItemsForVoucher

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, []string{helper.GroupIDSingleSku1})
	helper.setMockEmployeeDiscount10Percent(t, "employee-discount", []string{helper.GroupIDTwoSkus})
	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDTwoSkus}, true, false)

	// no employee discounts for regular customer
	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)
	// 5 CHF crossprice for Sku1 + 2 CHF voucher discount
	assert.Equal(t, 7.0, summary.TotalDiscountApplicable)

	// Change customer type to employee => additional employee discount
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	discounts, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)

	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)
	// 5 CHF crossprice for Sku1 + 0.5+2.0 employee discount + 1.8 CHF voucher discount (for Sku2)
	assert.Equal(t, 9.3, summary.TotalDiscountApplicable)
}
func TestCumulationEmployeeDiscountAndVoucherExcludeDiscountedItemsSAPCrossPrice(t *testing.T) {

	// Same as TestCumulationEmployeeDiscountAndVoucherExcludeDiscountedItems but with SAP cross-price instead of webshop cross-price

	// - Cart with 2 Items
	// - Cross-price on Sku1
	// - Employee discount on both items
	// - 10% Voucher
	// Expected result: Employee discount is granted regardless of flag ExcludeAlreadyDiscountedItemsForVoucher

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()
	// Simulate SAP cross price
	articleCollection.Articles[0].Price = 5.0
	articleCollection.Articles[0].AllowCrossPriceCalculation = false

	helper.setMockEmployeeDiscount10Percent(t, "employee-discount", []string{helper.GroupIDTwoSkus})
	voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDTwoSkus}, true, false)

	// no employee discounts for regular customer
	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)
	// 2 CHF voucher discount (Sku2)
	assert.Equal(t, 2.0, summary.TotalDiscountApplicable)

	// Change customer type to employee => additional employee discount
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	discounts, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.NoError(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)
	//  0.5+2.0 employee discount + 1.8 CHF voucher discount (for Sku2)
	assert.Equal(t, 4.3, summary.TotalDiscountApplicable)
}
