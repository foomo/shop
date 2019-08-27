package pricerule

import (
	"testing"

	"github.com/foomo/shop/utils"
	"github.com/stretchr/testify/assert"
)

type cumulationTestHelper struct{}

func (helper cumulationTestHelper) getMockArticleCollection() *ArticleCollection {
	return &ArticleCollection{
		Articles: []*Article{
			{
				ID:                         Sku1,
				Price:                      10.0,
				Quantity:                   1,
				AllowCrossPriceCalculation: true,
			},
			{
				ID:                         Sku2,
				Price:                      20.0,
				Quantity:                   1,
				AllowCrossPriceCalculation: true,
			},
		},
	}
}
func (helper cumulationTestHelper) setMockPriceRuleAndVoucher(t *testing.T, name string, amount float64, includedProductGroupIDS []string, excludeAlrreadyDiscountedItems bool) string {

	// PRICERULE
	priceRule := NewPriceRule("PriceRule-" + name)
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	priceRule.ExcludeAlreadyDiscountedForVoucher = excludeAlrreadyDiscountedItems
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

func TestCumulationTwoVouchers_OnePerSku(t *testing.T) {

	// Cart with 2 Items
	// Expected result: one of the vouchers is applied to each item
	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 10.0, []string{GroupIDSingleSku1}, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDSingleSku2}, false)

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

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 10.0, []string{GroupIDSingleSku1}, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDSingleSku1}, false)

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

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-1", 5.0, []string{GroupIDTwoSkus}, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-2", 10.0, []string{GroupIDTwoSkus}, false)

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

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0, []string{GroupIDTwoSkus})
	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 5.0, []string{GroupIDSingleSku1}, false)
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDSingleSku1}, false)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// Sku1: 5 + 0.5, Sku2: 5 => 10.5
	assert.Equal(t, 10.5, summary.TotalDiscountApplicable)

}
func TestCumulationProductPromo(t *testing.T) {

	// Cart with 2 Items
	// Expected result: - only better cross price should be applied

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 2.0, []string{GroupIDTwoSkus})
	helper.setMockPriceRuleCrossPrice(t, "crossprice2", 5.0, []string{GroupIDTwoSkus})

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, []string{}, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 5+5 (5 on each product) = 10 CHF
	assert.Equal(t, 10.0, summary.TotalDiscountApplicable)

}

func TestCumulationForExcludeVoucherOnCrossPriceWebhop(t *testing.T) {
	// Cart with 2 Items
	// Expected result: voucher is only applied to sku2. Sku1 is skipped due to existing webshop cross-price

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0, []string{GroupIDSingleSku1})
	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-1", 10.0, []string{GroupIDTwoSkus}, true)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// Sku1: 5, Sku2: 2
	assert.Equal(t, 7.0, summary.TotalDiscountApplicable)

}
func TestCumulationForExcludeVoucherOnCrossPriceSAP(t *testing.T) {

	// Cart with 2 Items
	// Expected result: voucher is only applied to sku2. Sku1 is skipped due to existing SAP cross-price
	// 					(indicated by AllowCrossPriceCalculation == false)

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()
	for i, article := range articleCollection.Articles {
		if i == 0 {
			article.AllowCrossPriceCalculation = false // this is false if there already is a SAP crossprice
		}
	}

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-1", 10.0, []string{GroupIDTwoSkus}, true)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1}, nil, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// Sku2: 2 CHF
	assert.Equal(t, 2.0, summary.TotalDiscountApplicable)

}
