package pricerule

import (
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
	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()
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
		group.AddGroupItemIDs(items)
		err := group.Upsert()
		if err != nil {
			t.Fatal(err, "Could not create product groups")
		}
	}
}

func (helper cumulationTestHelper) createMockCustomerGroups(t *testing.T, customerGroups []string) {
	for _, groupID := range customerGroups {
		group := new(Group)
		group.Type = CustomerGroup
		group.ID = groupID
		group.Name = groupID
		err := group.Upsert()
		if err != nil {
			t.Fatal(err, "Could not create customer groups")
		}
		group.AddGroupItemIDsAndPersist([]string{groupID})
	}
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
func (helper cumulationTestHelper) setMockEmployeeDiscount10Percent(t *testing.T, includedProductGroupIDS []string) {

	// PRICERULE
	priceRule := NewPriceRule("PriceRule-EmployeeDiscount")
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupEmployee}
	assert.Nil(t, priceRule.Upsert())

	return
}
func (helper cumulationTestHelper) setMockPriceRuleAndVoucherXPercent(t *testing.T, name string, amount float64, includedProductGroupIDS []string, excludeAlreadyDiscountedItems bool, excludeEmployees bool) string {

	// PRICERULE
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

	voucherCode := "voucherCode-" + priceRule.ID

	voucher := NewVoucher(voucherCode, voucherCode, priceRule, "")
	assert.Nil(t, voucher.Upsert())
	return voucherCode
}

func (helper cumulationTestHelper) setMockPriceRuleCrossPrice(t *testing.T, name string, amount float64, includedProductGroupIDS []string) {
	// PRICERULE 0
	priceRule := NewPriceRule("PriceRulePromotionProduct-" + name)
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	assert.Nil(t, priceRule.Upsert())
}

func (helper cumulationTestHelper) setMockPriceRuleAndVoucherXPercentAbsoluteCHF20(t *testing.T, cumulate bool) (string, string) {
	// PRICERULE
	priceRule := NewPriceRule("PriceRule-" + "CartAbsoluteCHF20")
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByAbsolute
	priceRule.Amount = 20.0
	priceRule.IncludedProductGroupIDS = []string{helper.GroupIDTwoSkus}
	priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupRegular}
	priceRule.ExcludeAlreadyDiscountedItemsForVoucher = false
	priceRule.CumulateWithOtherVouchers = cumulate
	assert.Nil(t, priceRule.Upsert())

	voucherCode := "voucherCode-" + priceRule.ID
	voucher := NewVoucher(voucherCode, voucherCode, priceRule, "")
	assert.Nil(t, voucher.Upsert())
	voucherCode2 := "voucherCode-" + priceRule.ID + "-2"
	voucher2 := NewVoucher(voucherCode2, voucherCode2, priceRule, "")
	assert.Nil(t, voucher2.Upsert())

	return voucherCode, voucherCode2
}
func (helper cumulationTestHelper) setMockPriceRuleAndVoucherXPercent10Percent(t *testing.T, cumulate bool) string {
	// PRICERULE
	priceRule := NewPriceRule("PriceRule-" + "Cart10Percent")
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10.0
	priceRule.IncludedProductGroupIDS = []string{helper.GroupIDTwoSkus}
	priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupRegular}
	priceRule.ExcludeAlreadyDiscountedItemsForVoucher = false
	priceRule.CumulateWithOtherVouchers = cumulate
	assert.Nil(t, priceRule.Upsert())

	voucherCode := "voucherCode-" + priceRule.ID

	voucher := NewVoucher(voucherCode, voucherCode, priceRule, "")
	assert.Nil(t, voucher.Upsert())
	return voucherCode
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
	assert.Nil(t, errApply)
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
	assert.Nil(t, errApply)
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
	assert.Nil(t, errApply)
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
	assert.Nil(t, errApply)
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
	assert.Nil(t, errApply)
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
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// helper.Sku1: 5, helper.Sku2: 2
	assert.Equal(t, 7.0, summary.TotalDiscountApplicable)

}
func TestCumulationForExcludeVoucherOnCrossPriceSAP(t *testing.T) {

	// Cart with 2 Items
	// Expected result: voucher is only applied to sku2. helper.Sku1 is skipped due to existing SAP cross-price
	// 					(indicated by AllowCrossPriceCalculation == false)

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
	assert.Nil(t, errApply)
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

	voucherCode1, _ := helper.setMockPriceRuleAndVoucherXPercentAbsoluteCHF20(t, true)
	voucherCode2 := helper.setMockPriceRuleAndVoucherXPercent10Percent(t, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
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

	voucherCode1, voucherCode2 := helper.setMockPriceRuleAndVoucherXPercentAbsoluteCHF20(t, true)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
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

	voucherCode1, voucherCode2 := helper.setMockPriceRuleAndVoucherXPercentAbsoluteCHF20(t, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
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
	setPriceRule := func(t *testing.T) (string, string) {
		// PRICERULE
		priceRule := NewPriceRule("PriceRule-" + "Bonus10CHF")
		priceRule.Description = priceRule.Name
		priceRule.Type = TypeBonusVoucher
		priceRule.Action = ActionCartByAbsolute
		priceRule.Amount = 10.0
		priceRule.IncludedCustomerGroupIDS = []string{helper.CustomerGroupRegular}

		priceRule.CalculateDiscountedOrderAmount = true
		priceRule.Priority = 999
		assert.Nil(t, priceRule.Upsert())

		voucherCode := "voucherCode-" + priceRule.ID
		voucher := NewVoucher(voucherCode, voucherCode, priceRule, "")
		assert.Nil(t, voucher.Upsert())
		voucherCode2 := "voucherCode-" + priceRule.ID + "-2"
		voucher2 := NewVoucher(voucherCode2, voucherCode2, priceRule, "")
		assert.Nil(t, voucher2.Upsert())

		return voucherCode, voucherCode2
	}

	articleCollection := helper.getMockArticleCollection()

	voucherCode1, voucherCode2 := setPriceRule(t)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
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

	// Get Discuonts for regular customer
	_, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.Nil(t, errApply)

	// 10% on 30 CHF
	assert.Equal(t, 3.0, summary.TotalDiscountApplicable)

	// Change customer type => no discount
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	_, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.Nil(t, errApply)

	// No discount for non-regulat customer
	assert.Equal(t, 0.0, summary.TotalDiscountApplicable)

}
func TestCumulationEmployeeDiscount(t *testing.T) {

	// Cart with 2 Items
	// Expected result: only employee customer gets discount

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	//voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDTwoSkus}, false, true)

	helper.setMockEmployeeDiscount10Percent(t, []string{helper.GroupIDTwoSkus})

	// no employee discuonts for regular customer
	_, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.Nil(t, errApply)

	assert.Equal(t, 0.0, summary.TotalDiscountApplicable)

	// Change customer type to eomplyee => discount applied
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	_, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.Nil(t, errApply)

	// No discount for non-regular customer
	assert.Equal(t, 3.0, summary.TotalDiscountApplicable)

}
func TestCumulationCrossPriceAndEmployeeDiscount(t *testing.T) {

	// Cart with 2 Items
	// Expected result: only employee customer gets discount

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	//voucherCode1 := helper.setMockPriceRuleAndVoucherXPercent(t, "voucher-sku1", 10.0, []string{helper.GroupIDTwoSkus}, false, true)
	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, []string{helper.GroupIDTwoSkus})
	helper.setMockEmployeeDiscount10Percent(t, []string{helper.GroupIDTwoSkus})

	// no employee discounts for regular customer
	_, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.Nil(t, errApply)

	// 5+5 CHF Crossprices, no employee discount
	assert.Equal(t, 10.0, summary.TotalDiscountApplicable)

	// Change customer type to eomplyee => discount applied
	articleCollection.CustomerType = helper.CustomerGroupEmployee
	_, summary, errApply = ApplyDiscounts(articleCollection, nil, []string{}, nil, 0.05, nil)
	assert.Nil(t, errApply)

	// 5+5 CHF Crossprices + 2 CHF employee discount
	assert.Equal(t, 12.0, summary.TotalDiscountApplicable)

}
