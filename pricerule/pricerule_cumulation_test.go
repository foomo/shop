package pricerule

import (
	"testing"

	"github.com/foomo/shop/utils"
	"github.com/stretchr/testify/assert"
)

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

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0, false, []string{helper.GroupIDTwoSkus})
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

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 2.0, false, []string{helper.GroupIDTwoSkus})
	helper.setMockPriceRuleCrossPrice(t, "crossprice2", 5.0, false, []string{helper.GroupIDTwoSkus})

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

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0, false, []string{helper.GroupIDSingleSku1})
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

	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, false, []string{helper.GroupIDTwoSkus})
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

	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, false, []string{helper.GroupIDSingleSku1})
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

	helper.setMockPriceRuleCrossPrice(t, "cross-price", 5.0, false, []string{helper.GroupIDSingleSku1})
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

func TestCrossPriceAndBuyXGetY(t *testing.T) {

	// Expected: Crossprice and BuyXPayY are both applied

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()
	allowCrossPriceCalculation := true
	// add 3rd article required for BuyXPayY Promo
	articleCollection.Articles = append(articleCollection.Articles, &Article{
		ID:                         helper.Sku3,
		Price:                      50.0,
		CrossPrice:                 50.0,
		Quantity:                   1,
		AllowCrossPriceCalculation: allowCrossPriceCalculation,
	})
	// adjuts prices
	articleCollection.Articles[0].Price = 100.0
	articleCollection.Articles[1].Price = 50.0

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 10.0, false, []string{helper.GroupIDThreeSkus})
	helper.setMockPriceRuleBuy3Pay2(t, "buyXPayY", []string{helper.GroupIDThreeSkus})

	tests := []struct {
		qty1                         float64
		qty2                         float64
		qty3                         float64
		expectedDiscountOrderService float64
		expectedDiscountCatalogue    float64
	}{
		{1.0, 1.0, 1.0, 70.0, 30.0}, // 10+10+10+10+40=70
		{2.0, 1.0, 1.0, 80.0, 40.0},
		{2.0, 2.0, 2.0, 140.0, 60.0}, // 2 items free
		{2.0, 1.0, 0.0, 70.0, 30.0},
		{0.0, 1.0, 1.0, 20.0, 20.0}, // only cross prices
	}

	for i, tt := range tests {
		// Order -------------------------------------------------------------------------------
		articleCollection.Articles[0].Quantity = tt.qty1
		articleCollection.Articles[1].Quantity = tt.qty2
		articleCollection.Articles[2].Quantity = tt.qty3

		// Calculation for orderservice
		_, summary, err := ApplyDiscounts(articleCollection, nil, []string{""}, []string{}, 0.05, nil)
		if err != nil {
			assert.NoError(t, err)
		}
		assert.Equal(t, tt.expectedDiscountOrderService, summary.TotalDiscount, "case orderservice", i)

		// Calculation for catalogue
		discountsCatalogue, _, err := ApplyDiscountsOnCatalog(articleCollection, nil, 0.05, nil)
		if err != nil {
			assert.NoError(t, err)
		}
		// Note: In catalalogue calculation summary is always empty, therefore we have to get the data directly from the discounts
		assert.Equal(t, tt.expectedDiscountCatalogue, helper.accumulateDiscountsOfItems(discountsCatalogue), "case catalogue", i)
	}
}
func TestCrossPriceAndEmployeeDiscount(t *testing.T) {

	// Expected: Crossprice and BuyXPayY are both applied

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()
	articleCollection.CustomerType = helper.CustomerGroupEmployee

	includedProdcuts := []string{helper.GroupIDTwoSkus}
	helper.setMockEmployeeDiscount10Percent(t, "employee-discount", includedProdcuts)

	tests := []struct {
		customerType                 string
		amountCrossPrice             float64
		isCrosspricePercent          bool
		expectedDiscountOrderService float64
		expectedDiscountCatalogue    float64
	}{
		{helper.CustomerGroupRegular, 5.0, false, 10.0, 10.0},
		{helper.CustomerGroupRegular, 10.0, true, 3.0, 3.0},
		{helper.CustomerGroupEmployee, 5.0, false, 12.0, 10.0},
		{helper.CustomerGroupEmployee, 10.0, true, 5.7, 3.0}, //@todo actual caculated result is 5.699999999999999 => fix rounding errors
	}

	for i, tt := range tests {
		articleCollection.CustomerType = tt.customerType
		helper.setMockPriceRuleCrossPrice(t, "crossprice1", tt.amountCrossPrice, tt.isCrosspricePercent, includedProdcuts)
		// Calculation for orderservice
		_, summary, err := ApplyDiscounts(articleCollection, nil, []string{""}, []string{}, 0.05, nil)
		if err != nil {
			assert.NoError(t, err)
		}

		assert.Equal(t, tt.expectedDiscountOrderService, utils.Round(summary.TotalDiscount, 2), "case orderservice", i)

		// Calculation for catalogue
		discountsCatalogue, _, err := ApplyDiscountsOnCatalog(articleCollection, nil, 0.05, nil)
		if err != nil {
			assert.NoError(t, err)
		}
		// Note: In catalalogue calculation summary is always empty, therefore we have to get the data directly from the discounts
		assert.Equal(t, tt.expectedDiscountCatalogue, helper.accumulateDiscountsOfItems(discountsCatalogue), "case catalogue", i)
	}
}
func TestQuantityThreshold(t *testing.T) {

	// Expected: Discount is applied if qty of one item is at least equal to threshold
	// Note: if threshold is met for ONE item, discount will be applied to ALL products eligible for this promo

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	tests := []struct {
		includedProducts []string
		amount           float64
		threshold        float64
		qty1             float64
		qty2             float64
		expectedDiscount float64
	}{
		{[]string{helper.GroupIDTwoSkus}, 1.0, 3.0, 1.0, 1.0, 0.0},
		{[]string{helper.GroupIDTwoSkus}, 1.0, 2.0, 1.0, 1.0, 2.0},
		{[]string{helper.GroupIDTwoSkus}, 1.0, 2.0, 2.0, 1.0, 3.0},
		{[]string{helper.GroupIDTwoSkus}, 1.0, 2.0, 2.0, 0.0, 2.0},
		{[]string{helper.GroupIDTwoSkus}, 1.0, 2.0, 2.0, 2.0, 4.0},
		{[]string{helper.GroupIDSingleSku1}, 1.0, 2.0, 2.0, 2.0, 2.0},
		{[]string{helper.GroupIDSingleSku1}, 1.0, 2.0, 1.0, 1.0, 0.0},
	}

	for i, tt := range tests {
		articleCollection.Articles[0].Quantity = tt.qty1
		articleCollection.Articles[1].Quantity = tt.qty2
		helper.setMockPriceRuleQtyThreshold(t, "qty-threshold", tt.amount, tt.threshold, tt.includedProducts)

		_, summary, err := ApplyDiscounts(articleCollection, nil, []string{""}, []string{}, 0.05, nil)
		if err != nil {
			assert.NoError(t, err)
		}

		assert.Equal(t, tt.expectedDiscount, summary.TotalDiscount, "case ", i)

	}
}
func TestThresholdAmount(t *testing.T) {

	// Expected: Discount is applied if threshold amount is met by all eligible items

	helper := newTesthelper()
	defer helper.cleanupTestData(t)
	helper.cleanupAndRecreateTestData(t)
	articleCollection := helper.getMockArticleCollection()

	tests := []struct {
		includedProducts []string
		thresholdAmount  float64
		qty1             float64
		qty2             float64
		expectedDiscount float64
	}{
		{[]string{helper.GroupIDTwoSkus}, 30.0, 1.0, 1.0, 10.0},
		{[]string{helper.GroupIDTwoSkus}, 40.0, 1.0, 1.0, 0.0},
		{[]string{helper.GroupIDTwoSkus}, 30.0, 0.0, 2.0, 10.0},
		{[]string{helper.GroupIDTwoSkus}, 30.0, 1.0, 1.0, 10.0},
		{[]string{helper.GroupIDTwoSkus}, 40.0, 1.0, 1.0, 0.0},
		{[]string{helper.GroupIDTwoSkus}, 30.0, 0.0, 2.0, 10.0},
	}

	for i, tt := range tests {
		articleCollection.Articles[0].Quantity = tt.qty1
		articleCollection.Articles[1].Quantity = tt.qty2
		helper.setMockPriceRuleThresholdAmount(t, "qty-threshold", 5.0, tt.thresholdAmount, tt.includedProducts)

		_, summary, err := ApplyDiscounts(articleCollection, nil, []string{""}, []string{}, 0.05, nil)
		if err != nil {
			assert.NoError(t, err)
		}

		assert.Equal(t, tt.expectedDiscount, summary.TotalDiscount, "case ", i)

	}
}
