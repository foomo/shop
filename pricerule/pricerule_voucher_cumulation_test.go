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
func (helper cumulationTestHelper) setMockPriceRuleAndVoucher(t *testing.T, name string, amount float64, includedProductGroupIDS []string) string {

	// PRICERULE
	priceRule := NewPriceRule("PriceRule-" + name)
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = includedProductGroupIDS
	assert.Nil(t, priceRule.Upsert())

	voucherCode := "voucherCode-" + priceRule.ID

	voucher := NewVoucher(voucherCode, voucherCode, priceRule, "")
	assert.Nil(t, voucher.Upsert())
	return voucherCode
}

func (helper cumulationTestHelper) setMockPriceRuleCrossPrice(t *testing.T, name string, amount float64) {
	// PRICERULE 0
	priceRule := NewPriceRule("PriceRulePromotionProduct-" + name)
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = amount
	priceRule.IncludedProductGroupIDS = []string{GroupIDTwoSkus}
	assert.Nil(t, priceRule.Upsert())
}

func TestCumulationTwoVouchers_OnePerSku(t *testing.T) {

	// Cart with 2 Items
	// Expected result: one of the vouchers is applied to each item
	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 10.0, []string{GroupIDSingleSku1})
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDSingleSku2})

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

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 10.0, []string{GroupIDSingleSku1})
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDSingleSku1})

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

	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 5.0, []string{GroupIDTwoSkus})
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDTwoSkus})

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 105 on 30 CHF
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

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 5.0)
	voucherCode1 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku1", 5.0, []string{GroupIDSingleSku1})
	voucherCode2 := helper.setMockPriceRuleAndVoucher(t, "voucher-sku2", 10.0, []string{GroupIDSingleSku1})

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{voucherCode1, voucherCode2}, nil, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// Sku1: 5 + 0.9, Sku2: 5 => 10.9
	assert.Equal(t, 10.5, summary.TotalDiscountApplicable)

}
func TestCumulationProductPromo(t *testing.T) {

	// Cart with 2 Items
	// Expected result: - only better cross price should be applied

	Init(t)

	helper := cumulationTestHelper{}
	articleCollection := helper.getMockArticleCollection()

	helper.setMockPriceRuleCrossPrice(t, "crossprice1", 2.0)
	helper.setMockPriceRuleCrossPrice(t, "crossprice2", 5.0)

	discounts, summary, errApply := ApplyDiscounts(articleCollection, nil, []string{}, []string{}, 0.05, nil)
	assert.Nil(t, errApply)
	utils.PrintJSON(discounts)
	utils.PrintJSON(summary)

	// 5+5 (5 on each product) = 10 CHF
	assert.Equal(t, 10.0, summary.TotalDiscountApplicable)

}
