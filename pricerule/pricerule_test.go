package pricerule

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const (
	GroupIDSale   = "Sale"
	GroupIDNormal = "Products"
	GroupIDShirts = "Shirts"

	PriceRuleIDSale         = "PriceRuleSale"
	PriceRuleIDSaleProduct  = "PriceRuleSaleProduct"
	PriceRuleIDSaleCustomer = "PriceRuleSaleCustomer"
	PriceRuleIDSaleVoucher  = "PriceRuleSaleVoucher"
	PriceRuleIDVoucher      = "PriceRuleVoucher"
	PriceRuleIDPayment      = "PriceRulePayment"

	VoucherID1   = "voucher1"
	VoucherCode1 = "voucher-code-1"

	VoucherID2   = "voucher2"
	VoucherCode2 = "voucher-code-2"
	// Products
	ProductID1 = "product-1"
	ProductID2 = "product-2"
	ProductID3 = "product-3"
	ProductID4 = "product-4"
	ProductID5 = "product-5"

	//SKUs

	ProductID1SKU1 = "product-1-sku1"
	ProductID1SKU2 = "product-1-sku2"

	ProductID2SKU1 = "product-2-sku1"
	ProductID2SKU2 = "product-2-sku2"

	ProductID3SKU1 = "product-3-sku1"
	ProductID3SKU2 = "product-3-sku2"

	ProductID4SKU1 = "product-4-sku1"
	ProductID4SKU2 = "product-4-sku2"

	ProductID5SKU1 = "product-5-sku1"
	ProductID5SKU2 = "product-5-sku2"

	PaymentMethodID1 = "PaymentMethodID1"
	PaymentMethodID2 = "PaymentMethodID2"

	CustomerID1 = "CustomerID1"
	CustomerID2 = "CustomerID2"

	CustomerGroupID1 = "CustomerGroupID1 - super customer"
	CustomerGroupID2 = "CustomerGroupID2 - employee"
)

var productsInGroups map[string][]string


func Init(t *testing.T) {

	productsInGroups = make(map[string][]string)
	productsInGroups[GroupIDSale] = []string{ProductID1, ProductID2, ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2}
	productsInGroups[GroupIDNormal] = []string{ProductID4, ProductID5, ProductID4SKU1, ProductID4SKU2, ProductID5SKU1, ProductID5SKU2}
	productsInGroups[GroupIDShirts] = []string{ProductID3, ProductID4, ProductID5, ProductID3SKU1, ProductID4SKU1, ProductID5SKU1, ProductID3SKU2, ProductID4SKU2, ProductID5SKU2}

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()
	checkGroupsNotExists(t)
	createMockCustomerGroups(t)
	createMockProductGroups(t)
	checkGroupsExists(t)

	createMockPriceRules(t)
	checkPriceRulesExists(t)

	createMockVouchers(t)
	checkVouchersExists(t)
}

// Test groups creation
func testGetApplicableVouchers(t *testing.T) {
	//remove all and add again
	productsInGroups = make(map[string][]string)
	productsInGroups[GroupIDSale] = []string{ProductID1, ProductID2, ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2}
	productsInGroups[GroupIDNormal] = []string{ProductID4, ProductID5, ProductID4SKU1, ProductID4SKU2, ProductID5SKU1, ProductID5SKU2}
	productsInGroups[GroupIDShirts] = []string{ProductID3, ProductID4, ProductID5, ProductID3SKU1, ProductID4SKU1, ProductID5SKU1, ProductID3SKU2, ProductID4SKU2, ProductID5SKU2}

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()
	checkGroupsNotExists(t)
	createMockCustomerGroups(t)
	createMockProductGroups(t)
	checkGroupsExists(t)
	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}

	// VOUCHERS ------------``
	priceRule := NewPriceRule(PriceRuleIDSaleVoucher)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSaleVoucher,
		"fr": PriceRuleIDSaleVoucher,
		"it": PriceRuleIDSaleVoucher,
	}
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 20.0
	priceRule.Priority = 800
	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}
	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule, err = GetPriceRuleByID(PriceRuleIDSaleVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher := NewVoucher(VoucherID1, VoucherCode1, priceRule, "")

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// ------------

	priceRule = NewPriceRule(PriceRuleIDVoucher)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDVoucher,
		"fr": PriceRuleIDVoucher,
		"it": PriceRuleIDVoucher,
	}
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10.0
	priceRule.Priority = 80
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.IncludedCustomerGroupIDS = []string{}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule, err = GetPriceRuleByID(PriceRuleIDVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher = NewVoucher(VoucherID2, VoucherCode2, priceRule, CustomerID2)

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// PRICERULES --------------------------------------------------------------------------------------
	applicableRules, err := PickApplicableVouchers([]string{VoucherCode1, VoucherCode2}, orderVo, []string{}, nil)
	spew.Dump(applicableRules)

}

func testShipping1(t *testing.T) {

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = ProductGroup
	group.ID = "shipping"
	group.Name = "shipping"
	group.AddGroupItemIDs([]string{"shipping-item-id"})
	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	group = new(Group)
	group.Type = CustomerGroup
	group.ID = "customer-group"
	group.Name = "customer"
	group.AddGroupItemIDs([]string{CustomerID1})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "product1-group"
	group.Name = "product1-group"
	group.AddGroupItemIDs([]string{"product1"})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert  product1 group ")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "product2-group"
	group.Name = "product2-group"
	group.AddGroupItemIDs([]string{"product2"})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert product2 group ")
	}

	// --------------------------------------------------------
	//create pricerule
	priceRule := NewPriceRule("shipping")
	priceRule.Name = map[string]string{
		"de": "shipping",
		"fr": "shipping",
		"it": "shipping",
	}
	priceRule.Type = TypeShipping
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 100
	priceRule.MinOrderAmount = 311.10 - 4.9
	priceRule.CalculateDiscountedOrderAmount = true
	priceRule.ExcludedItemIDsFromOrderAmountCalculation = []string{"shipping-item-id"}
	priceRule.IncludedProductGroupIDS = []string{"shipping"}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.Upsert()

	//customer promo 1

	priceRule = NewPriceRule("customer-promo30")
	priceRule.Name = map[string]string{
		"de": "customer-promo30",
		"fr": "customer-promo30",
		"it": "customer-promo30",
	}
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 30
	//priceRule.ExcludedProductGroupIDS = []string{"shipping"}
	priceRule.IncludedProductGroupIDS = []string{"product1-group", "shipping"}
	priceRule.IncludedCustomerGroupIDS = []string{"customer-group"}
	priceRule.Upsert()

	priceRule = NewPriceRule("customer-promo15")
	priceRule.Name = map[string]string{
		"de": "customer-promo15",
		"fr": "customer-promo15",
		"it": "customer-promo15",
	}
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 15
	//priceRule.ExcludedProductGroupIDS = []string{"shipping"}
	priceRule.IncludedProductGroupIDS = []string{"product2-group"}
	priceRule.IncludedCustomerGroupIDS = []string{"customer-group"}
	priceRule.Upsert()

	priceRule = NewPriceRule("product2-promo")
	priceRule.Name = map[string]string{
		"de": "product2-promo",
		"fr": "product2-promo",
		"it": "product2-promo",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10

	//priceRule.ExcludedProductGroupIDS = []string{"shipping"}
	priceRule.IncludedProductGroupIDS = []string{"product2-group"}
	priceRule.IncludedCustomerGroupIDS = []string{"customer-group"}
	priceRule.Upsert()

	//create order

	// order = articleCollection
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1
	orderVo.CustomerType = CustomerID1
	positionVo := &Article{}
	positionVo.ID = "shipping-item-id"
	positionVo.Price = 4.90
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = "product1"
	positionVo.Price = 399.0
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = "product2"
	positionVo.Price = 29.90
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{""}, []string{}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testBlacklist(t *testing.T) {
	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = ProductGroup
	group.ID = "sale"
	group.Name = "sale"
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2})
	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}
	//create pricerule
	priceRule := NewPriceRule("sale")
	priceRule.Name = map[string]string{
		"de": "sale",
		"fr": "sale",
		"it": "sale",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{"sale"}
	priceRule.Upsert()

	//create pricerule with blacklist

	//create blacklist
	group = new(Group)
	group.Type = BlacklistGroup
	group.ID = "blacklist"
	group.Name = "blacklist"
	group.AddGroupItemIDs([]string{ProductID3SKU1})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert blacklist product group ")
	}

	group = new(Group)
	group.Type = BlacklistGroup
	group.ID = "blacklist1"
	group.Name = "blacklist1"
	group.AddGroupItemIDs([]string{ProductID2SKU1, ProductID1SKU2})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert blacklist product group ")
	}

	// Order -------------------------------------------------------------------------------
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID2SKU1
	positionVo.Price = 300
	positionVo.Quantity = 2
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 500
	positionVo.Quantity = 5
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 100
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, []string{PaymentMethodID1}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testDiscountDistribution(t *testing.T) {
	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = CustomerGroup
	group.ID = "employees"
	group.Name = "employees"
	group.AddGroupItemIDs([]string{"employeeID"})

	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert employees")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "products1"
	group.Name = "products1"
	group.AddGroupItemIDs([]string{"product1"})

	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert products1")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "products2"
	group.Name = "products2"
	group.AddGroupItemIDs([]string{"product2"})

	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert products2")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "vouchergroup"
	group.Name = "vouchergroup"
	group.AddGroupItemIDs([]string{})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert vouchergroup")
	}

	//create pricerule
	priceRule := NewPriceRule("ruleproducts1")
	priceRule.Name = map[string]string{
		"de": "ruleproducts1",
		"fr": "ruleproducts1",
		"it": "ruleproducts1",
	}
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = 25
	priceRule.MinOrderAmount = 0
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.IncludedProductGroupIDS = []string{"products1"}
	priceRule.Upsert()

	//create pricerule
	priceRule = NewPriceRule("ruleproducts2")
	priceRule.Name = map[string]string{
		"de": "ruleproducts2",
		"fr": "ruleproducts2",
		"it": "ruleproducts2",
	}
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = 2
	priceRule.MinOrderAmount = 0
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.IncludedProductGroupIDS = []string{"products2"}
	priceRule.Upsert()

	priceRule = NewPriceRule("voucher")
	priceRule.Name = map[string]string{
		"de": "voucher",
		"fr": "voucher",
		"it": "voucher",
	}
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10
	priceRule.MinOrderAmount = 0
	priceRule.Upsert()

	voucher := NewVoucher("voucherID", "vouchercode", priceRule, "")
	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// Order

	orderVo := &ArticleCollection{}
	orderVo.CustomerID = "employeeID"

	positionVo := &Article{}
	positionVo.ID = "product1"
	positionVo.Price = 99.90
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = "product2"
	positionVo.Price = 19.90
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{"vouchercode"}, []string{}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testBestOption(t *testing.T) {
	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = ProductGroup
	group.ID = "group1"
	group.Name = "group1"
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2})
	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "group2"
	group.Name = "group2"
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	group = new(Group)
	group.Type = ProductGroup
	group.ID = "shipping"
	group.Name = "shipping"
	group.AddGroupItemIDs([]string{"shipping"})
	err = group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	//create pricerule
	priceRule := NewPriceRule("rule-group1")
	priceRule.Name = map[string]string{
		"de": "rule-group1",
		"fr": "rule-group1",
		"it": "rule-group1",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 60
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{"group1"}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.ExcludedProductGroupIDS = []string{"shipping"}
	priceRule.CheckoutAttributes = []string{}
	priceRule.QtyThreshold = 3.0
	priceRule.Upsert()

	//create pricerule
	priceRule = NewPriceRule("rule-group2")
	priceRule.Name = map[string]string{
		"de": "rule-group2",
		"fr": "rule-group2",
		"it": "rule-group2",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 50
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{"group2"}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.ExcludedProductGroupIDS = []string{}
	priceRule.CheckoutAttributes = []string{}
	priceRule.QtyThreshold = 0
	priceRule.Upsert()

	//create pricerule
	priceRule = NewPriceRule("rule-group3")
	priceRule.Name = map[string]string{
		"de": "rule-group3",
		"fr": "rule-group3",
		"it": "rule-group3",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 5
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.ExcludedProductGroupIDS = []string{"shipping"}
	priceRule.CheckoutAttributes = []string{}
	priceRule.QtyThreshold = 0
	priceRule.Upsert()
	//create pricerule

	// Order -------------------------------------------------------------------------------
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 2
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID2SKU1
	positionVo.Price = 200
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 100
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = "shipping"
	positionVo.Price = 5
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, []string{}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testDiscountFoItemSets(t *testing.T) {
	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = ProductGroup
	group.ID = "sale"
	group.Name = "sale"
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2})
	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}
	//create pricerule
	priceRule := NewPriceRule("itemset-discount")
	priceRule.Name = map[string]string{
		"de": "itemset-discount",
		"fr": "itemset-discount",
		"it": "itemset-discount",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemSetAbsolute
	priceRule.Amount = 10
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.CheckoutAttributes = []string{}
	priceRule.ItemSets = [][]string{
		[]string{ProductID1SKU1, ProductID1SKU2},
		[]string{ProductID2SKU1, ProductID2SKU2},
	}
	priceRule.QtyThreshold = 0
	priceRule.Upsert()

	//create pricerule

	// Order -------------------------------------------------------------------------------
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID2SKU1
	positionVo.Price = 300
	positionVo.Quantity = 2
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 500
	positionVo.Quantity = 5
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 100
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, []string{PaymentMethodID1}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testVoucherRuleWithCheckoutAttributes(t *testing.T) {
	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = ProductGroup
	group.ID = "sale"
	group.Name = "sale"
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID1SKU2})
	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	//create pricerule
	priceRule := NewPriceRule(PriceRuleIDSale)
	priceRule.Name = map[string]string{
		"de": "normal-discount",
		"fr": "normal-discount",
		"it": "normal-discount",
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = 10
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{group.ID}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.CheckoutAttributes = []string{}
	priceRule.QtyThreshold = 0
	priceRule.Upsert()

	//create pricerule

	//create voucher

	voucher := NewVoucher(VoucherID1, VoucherCode1, priceRule, "")
	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// Order -------------------------------------------------------------------------------
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID1SKU2
	positionVo.Price = 300
	positionVo.Quantity = float64(1)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 500
	positionVo.Quantity = float64(1)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, []string{PaymentMethodID1}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testShipping(t *testing.T) {

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	group := new(Group)
	group.Type = ProductGroup
	group.ID = "shipping"
	group.Name = "shipping"
	group.AddGroupItemIDs([]string{"shipping-item-id"})
	err := group.Upsert()
	if err != nil {
		t.Fatal("Could not upsert shipping product group ")
	}

	//create pricerule
	priceRule := NewPriceRule("shipping")
	priceRule.Name = map[string]string{
		"de": "shipping",
		"fr": "shipping",
		"it": "shipping",
	}
	priceRule.Type = TypeShipping
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByPercent
	priceRule.Amount = 5
	priceRule.MinOrderAmount = 50
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.CalculateDiscountedOrderAmount = true
	priceRule.ExcludedItemIDsFromOrderAmountCalculation = []string{"shipping-item-id"}
	priceRule.IncludedProductGroupIDS = []string{"shipping"}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.Upsert()

	priceRule = NewPriceRule("general-promotion")
	priceRule.Name = map[string]string{
		"de": "general-promotio",
		"fr": "general-promotio",
		"it": "general-promotio",
	}
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = 50
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.CalculateDiscountedOrderAmount = false
	priceRule.ExcludedProductGroupIDS = []string{"shipping"}
	priceRule.Upsert()

	//create order

	// order = articleCollection
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1
	positionVo := &Article{}
	positionVo.ID = "shipping-item-id"
	positionVo.Price = 5.0
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = "normal-item-id"
	positionVo.Price = 100.0
	positionVo.Quantity = 1
	orderVo.Articles = append(orderVo.Articles, positionVo)

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{""}, []string{}, 0.05, nil)
	spew.Dump(discountsVo, summary, err)

}

func testCache(t *testing.T) {
	Init(t)

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()

	//cache.InitCatalogCalculationCache()
	//spew.Dump(cache.GetGroupsCache())
	//cache.ClearCatalogCalculationCache()

	for _, groupID := range []string{GroupIDSale, GroupIDNormal, GroupIDShirts} {
		group := new(Group)
		group.Type = ProductGroup
		group.ID = groupID
		group.Name = groupID
		group.AddGroupItemIDs(productsInGroups[groupID])
		err := group.Upsert()
		if err != nil {
			t.Fatal("Could not upsert product group " + groupID)
		}
	}

	groupID := "ProductsToScale"
	//createGroup
	group := new(Group)
	group.Type = ProductGroup
	group.ID = groupID
	group.Name = groupID

	//create pricerule
	priceRule := NewPriceRule(PriceRuleIDSale)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByAbsolute
	priceRule.Amount = 10
	priceRule.IncludedProductGroupIDS = []string{groupID}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.Upsert()
	//insert as well
	group.AddGroupItemIDs([]string{ProductID1SKU1})

	err := group.Upsert()
	if err != nil {
		log.Println(err)
	}

	cache.InitCatalogCalculationCache()
	//spew.Dump(cache.GetGroupsCache())

	//create articleCollection
	orderVo, err := createMockOrderScaled(t)
	if err != nil {
		panic(err)
	}
	discountsVo, summary, err := ApplyDiscountsOnCatalog(orderVo, nil, 0.05, nil)
	spew.Dump(discountsVo, summary, err)
}

// Test groups creation
func testScaled(t *testing.T) {
	//Init

	//Init
	RemoveAllGroups()
	RemoveAllPriceRules()

	// Create group --------------------------------------------------------------------------
	groupID := "productstoscale"
	//createGroup
	group := new(Group)
	group.Type = ProductGroup
	group.ID = groupID
	group.Name = groupID
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID3SKU2})
	group.Upsert()

	// Create pricerule ----------------------------------------------------------------------
	priceRule := NewPriceRule(PriceRuleIDSale)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name

	priceRule.Action = ActionScaled
	priceRule.ScaledAmounts = append(priceRule.ScaledAmounts, ScaledAmountLevel{FromValue: 2.0, ToValue: 10.0, Amount: 10, IsScaledAmountPercentage: true, IsFromToPrice: false})

	priceRule.MaxUses = 10
	priceRule.MaxUsesPerCustomer = 10
	priceRule.IncludedProductGroupIDS = []string{"productstoscale"}
	priceRule.IncludedCustomerGroupIDS = []string{}

	err := priceRule.Upsert()
	if err != nil {
		panic(err)
	}
	// Order -------------------------------------------------------------------------------
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 2
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID1SKU2
	positionVo.Price = 300
	positionVo.Quantity = float64(2)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 500
	positionVo.Quantity = float64(2)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	// Order -------------------------------------------------------------------------------
	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{""}, []string{}, 0.05, nil)
	timeTrack(now, "Apply scaled voucher")
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}
	fmt.Println("discounts for scaled percentage")
	spew.Dump(discountsVo)
	spew.Dump(*summary)
}

// Test groups creation
func testBuyXGetY(t *testing.T) {
	//Init
	RemoveAllGroups()
	RemoveAllPriceRules()

	//create group --------------------------------------------------------------------

	groupID := "discounted"
	group := new(Group)
	group.Type = ProductGroup
	group.ID = groupID
	group.Name = groupID
	group.AddGroupItemIDs([]string{ProductID1SKU1, ProductID1SKU2, ProductID2SKU1})
	group.Upsert()

	//create pricerule --------------------------------------------------------------------
	priceRule := NewPriceRule(PriceRuleIDSale)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionBuyXPayY
	priceRule.X = 3
	priceRule.Y = 1
	priceRule.WhichXYFree = XYMostExpensiveFree
	priceRule.WhichXYList = []string{ProductID2SKU1, ProductID1SKU1, ProductID1SKU2}
	priceRule.MaxUses = 10
	priceRule.MaxUsesPerCustomer = 10
	priceRule.IncludedProductGroupIDS = []string{"discounted"}
	priceRule.IncludedCustomerGroupIDS = []string{}

	err := priceRule.Upsert()
	if err != nil {
		panic(err)
	}
	// Order -------------------------------------------------------------------------------
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 2
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID1SKU2
	positionVo.Price = 300
	positionVo.Quantity = float64(2)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID2SKU1
	positionVo.Price = 500
	positionVo.Quantity = float64(2)
	orderVo.Articles = append(orderVo.Articles, positionVo)
	// Order -------------------------------------------------------------------------------
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{""}, []string{}, 0.05, nil)
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	fmt.Println("discounts for buy x get y")
	spew.Dump(discountsVo)
	spew.Dump(*summary)

}

// Test groups creation
func testExclude(t *testing.T) {
	//remove all and add again
	productsInGroups = make(map[string][]string)
	productsInGroups[GroupIDSale] = []string{ProductID1, ProductID2, ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2}
	productsInGroups[GroupIDNormal] = []string{ProductID4, ProductID5, ProductID4SKU1, ProductID4SKU2, ProductID5SKU1, ProductID5SKU2}
	productsInGroups[GroupIDShirts] = []string{ProductID3, ProductID4, ProductID5, ProductID3SKU1, ProductID4SKU1, ProductID5SKU1, ProductID3SKU2, ProductID4SKU2, ProductID5SKU2}

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()
	checkGroupsNotExists(t)
	createMockCustomerGroups(t)
	createMockProductGroups(t)
	checkGroupsExists(t)
	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}

	priceRule := NewPriceRule(PriceRuleIDSaleProduct)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSaleProduct,
		"fr": PriceRuleIDSaleProduct,
		"it": PriceRuleIDSaleProduct,
	}
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByPercent
	priceRule.Amount = 10.0
	priceRule.Priority = 90
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.ExcludedProductGroupIDS = []string{GroupIDSale}
	priceRule.IncludedCustomerGroupIDS = []string{}
	priceRule.MinOrderAmount = 100
	priceRule.MinOrderAmountApplicableItemsOnly = true
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	productGroupIDsPerPosition := getProductGroupIDsPerPosition(orderVo, false)
	calculationParameters := &CalculationParameters{}
	calculationParameters.productGroupIDsPerPosition = productGroupIDsPerPosition
	calculationParameters.isCatalogCalculation = false
	calculationParameters.articleCollection = orderVo
	//spew.Dump(productGroupIDsPerPosition)
	for _, article := range orderVo.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRule, article, calculationParameters, nil)
		log.Println(article.ID + " " + priceRule.ID + " " + strconv.FormatBool(ok))
	}

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, []string{"blah"}, 0.05, nil)
	spew.Dump(discountsVo)
	spew.Dump(*summary)

}

// Test groups creation
func testMaxOrder(t *testing.T) {
	//remove all and add again
	productsInGroups = make(map[string][]string)
	productsInGroups[GroupIDSale] = []string{ProductID1, ProductID2, ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2}
	productsInGroups[GroupIDNormal] = []string{ProductID4, ProductID5, ProductID4SKU1, ProductID4SKU2, ProductID5SKU1, ProductID5SKU2}
	productsInGroups[GroupIDShirts] = []string{ProductID3, ProductID4, ProductID5, ProductID3SKU1, ProductID4SKU1, ProductID5SKU1, ProductID3SKU2, ProductID4SKU2, ProductID5SKU2}

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()
	checkGroupsNotExists(t)
	createMockCustomerGroups(t)
	createMockProductGroups(t)
	checkGroupsExists(t)
	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}

	// PRICERULES --------------------------------------------------------------------------------------
	//Customer price rule

	priceRule := NewPriceRule(PriceRuleIDSaleCustomer)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByPercent
	priceRule.Amount = 10.0
	priceRule.Priority = 90
	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}
	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}
	priceRule.MinOrderAmount = 0
	priceRule.MinOrderAmountApplicableItemsOnly = true
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	// VOUCHERS ------------
	priceRule = NewPriceRule(PriceRuleIDSaleVoucher)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSaleVoucher,
		"fr": PriceRuleIDSaleVoucher,
		"it": PriceRuleIDSaleVoucher,
	}
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 20.0
	priceRule.Priority = 800
	priceRule.MinOrderAmount = 0
	priceRule.ValidFrom = time.Date(1999, 12, 1, 12, 0, 0, 0, time.UTC)
	priceRule.ValidTo = time.Date(2016, 12, 1, 12, 0, 0, 0, time.UTC)

	priceRule.MinOrderAmountApplicableItemsOnly = false
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.IncludedCustomerGroupIDS = []string{}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule, err = GetPriceRuleByID(PriceRuleIDSaleVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher := NewVoucher(VoucherID1, VoucherCode1, priceRule, "")

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// PRICERULES --------------------------------------------------------------------------------------
	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode1}, []string{PaymentMethodID1}, 0.05, nil)
	timeTrack(now, "Apply multiple price rules")
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	fmt.Println("discounts")
	spew.Dump(discountsVo)
	spew.Dump(*summary)

	//validate the voucher

	_, message := ValidateVoucher(VoucherCode1, orderVo, []string{})
	fmt.Println(message)

}

// Test groups creation
func testTwoStepWorkflow(t *testing.T) {
	//remove all and add again
	productsInGroups = make(map[string][]string)
	productsInGroups[GroupIDSale] = []string{ProductID1, ProductID2, ProductID1SKU1, ProductID1SKU2, ProductID2SKU1, ProductID2SKU2}
	productsInGroups[GroupIDNormal] = []string{ProductID4, ProductID5, ProductID4SKU1, ProductID4SKU2, ProductID5SKU1, ProductID5SKU2}
	productsInGroups[GroupIDShirts] = []string{ProductID3, ProductID4, ProductID5, ProductID3SKU1, ProductID4SKU1, ProductID5SKU1, ProductID3SKU2, ProductID4SKU2, ProductID5SKU2}

	RemoveAllGroups()
	RemoveAllPriceRules()
	RemoveAllVouchers()
	checkGroupsNotExists(t)
	createMockCustomerGroups(t)
	createMockProductGroups(t)
	checkGroupsExists(t)
	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}

	// PRICERULES --------------------------------------------------------------------------------------
	//Customer price rule

	priceRule := NewPriceRule(PriceRuleIDSaleCustomer)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionCustomer
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByPercent
	priceRule.Amount = 10.0
	priceRule.Priority = 90
	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}
	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule = NewPriceRule(PriceRuleIDSaleProduct)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionProduct
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionCartByPercent
	priceRule.Amount = 10.0
	priceRule.Priority = 100
	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}
	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	// VOUCHERS ------------``
	priceRule = NewPriceRule(PriceRuleIDSaleVoucher)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSaleVoucher,
		"fr": PriceRuleIDSaleVoucher,
		"it": PriceRuleIDSaleVoucher,
	}
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 20.0
	priceRule.Priority = 800
	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}
	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule, err = GetPriceRuleByID(PriceRuleIDSaleVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher := NewVoucher(VoucherID1, VoucherCode1, priceRule, "")

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// ------------

	priceRule = NewPriceRule(PriceRuleIDVoucher)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDVoucher,
		"fr": PriceRuleIDVoucher,
		"it": PriceRuleIDVoucher,
	}
	priceRule.Type = TypeVoucher
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionItemByPercent
	priceRule.Amount = 10.0
	priceRule.Priority = 80
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.IncludedCustomerGroupIDS = []string{}
	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule, err = GetPriceRuleByID(PriceRuleIDVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher = NewVoucher(VoucherID2, VoucherCode2, priceRule, CustomerID2)

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	// PRICERULES --------------------------------------------------------------------------------------
	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode2, VoucherCode1}, []string{PaymentMethodID1}, 0.05, nil)
	timeTrack(now, "Apply multiple price rules")
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	fmt.Println("discounts")
	spew.Dump(discountsVo)
	spew.Dump(*summary)

}

// Test groups creation
func testPricerulesWorkflow(t *testing.T) {
	//remove all and add again
	Init(t)

	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}
	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode1}, []string{PaymentMethodID1}, 0.05, nil)
	timeTrack(now, "Apply multiple price rules")
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	fmt.Println("discounts")
	spew.Dump(discountsVo)
	spew.Dump(*summary)

}

// Test checkout functionality
func testCheckoutWorkflow(t *testing.T) {
	//remove all and add again
	Init(t)

	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}
	now := time.Now()
	discountsVo, _, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode1}, []string{PaymentMethodID1}, 0.05, nil)
	timeTrack(now, "Apply multiple price rules")
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	now = time.Now()
	ok, reason := ValidateVoucher(VoucherCode1, orderVo, []string{})
	if !ok {
		log.Println("VOUCHER INVALID" + VoucherCode1 + reason)
	}
	timeTrack(now, "Validated voucher")

	now = time.Now()
	err = CommitDiscounts(&discountsVo, orderVo.CustomerID)
	err = CommitDiscounts(&discountsVo, orderVo.CustomerID)
	if err != nil {
		log.Println("Already redeemed")
	}
	timeTrack(now, "Order committed")

}

// GROUPS -----------------------------------
func checkGroupsNotExists(t *testing.T) {
	for _, groupID := range []string{GroupIDSale, GroupIDNormal, GroupIDShirts} {
		group, _ := GetGroupByID(groupID, nil)
		if group != nil {
			t.Error("Group " + groupID + " should NOT exist after deletion")
		}
	}
}

func checkGroupsExists(t *testing.T) {
	for _, groupID := range []string{GroupIDSale, GroupIDNormal, GroupIDShirts} {
		group, _ := GetGroupByID(groupID, nil)
		if group == nil {
			t.Error("Group " + groupID + " should EXIST after creation")
		}
	}
}

func createMockProductGroups(t *testing.T) {
	for _, groupID := range []string{GroupIDSale, GroupIDNormal, GroupIDShirts} {
		group := new(Group)
		group.Type = ProductGroup
		group.ID = groupID
		group.Name = groupID
		group.AddGroupItemIDs(productsInGroups[groupID])
		err := group.Upsert()
		if err != nil {
			t.Fatal("Could not upsert product group " + groupID)
		}
	}
}

func createMockCustomerGroups(t *testing.T) {
	for _, groupID := range []string{CustomerGroupID1, CustomerGroupID2} {
		group := new(Group)
		group.Type = CustomerGroup
		group.ID = groupID
		group.Name = groupID
		group.AddGroupItemIDs([]string{CustomerID1})
		err := group.Upsert()

		if err != nil {
			log.Println(err)
			t.Fatal("Could not upsert customer group " + groupID)
		}
		group.AddGroupItemIDsAndPersist([]string{CustomerID2})
	}
}

// PRICERULES ---------------------------------------------

func createMockPriceRules(t *testing.T) {

	//Sale price rule

	priceRule := NewPriceRule(PriceRuleIDSale)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionOrder

	priceRule.Description = priceRule.Name

	priceRule.Action = ActionCartByPercent

	priceRule.Amount = 10.0

	priceRule.Priority = 10

	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}

	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}

	err := priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	//Sale price rule

	priceRule = NewPriceRule(PriceRuleIDSaleVoucher)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSaleVoucher,
		"fr": PriceRuleIDSaleVoucher,
		"it": PriceRuleIDSaleVoucher,
	}
	priceRule.Type = TypeVoucher

	priceRule.Description = priceRule.Name

	priceRule.Action = ActionItemByPercent
	priceRule.Priority = 9
	priceRule.Amount = 30.0

	priceRule.IncludedProductGroupIDS = []string{GroupIDSale}

	priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID2}

	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	//Voucher price rule

	priceRule = NewPriceRule(PriceRuleIDPayment)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDPayment,
		"fr": PriceRuleIDPayment,
		"it": PriceRuleIDPayment,
	}
	priceRule.Priority = 99
	priceRule.Type = TypePaymentMethodDiscount

	priceRule.Description = priceRule.Name

	priceRule.Action = ActionItemByPercent

	priceRule.Amount = 2.0 //2 ActionCartByPercent

	priceRule.CheckoutAttributes = []string{PaymentMethodID1}

	priceRule.IncludedProductGroupIDS = []string{} //for all products

	err = priceRule.Upsert()
	if err != nil {
		panic(err)
	}
}

func checkPriceRulesExists(t *testing.T) {
	for _, ID := range []string{PriceRuleIDSale, PriceRuleIDSaleVoucher, PriceRuleIDPayment} {
		priceRule, _ := GetPriceRuleByID(ID, nil)
		if priceRule == nil {
			t.Error("PriceRule " + ID + " should EXIST after creation")
		}
	}
}

func createMockVouchers(t *testing.T) {
	priceRule, err := GetPriceRuleByID(PriceRuleIDSaleVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher := NewVoucher(VoucherID1, VoucherCode1, priceRule, "")

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}

	priceRule, err = GetPriceRuleByID(PriceRuleIDSaleVoucher, nil)
	if err != nil {
		panic(err)
	}
	voucher = NewVoucher(VoucherID2, VoucherCode2, priceRule, CustomerID2)

	err = voucher.Upsert()
	if err != nil {
		panic(err)
	}
}

func checkVouchersExists(t *testing.T) {
	for _, ID := range []string{VoucherID1, VoucherID2} {
		voucher, _ := GetVoucherByID(ID, nil)
		if voucher == nil {
			t.Error("Voucher " + ID + " should EXIST after creation")
		}
	}
}

func createMockOrder(t *testing.T) (*ArticleCollection, error) {
	orderVo := &ArticleCollection{}

	orderVo.CustomerID = CustomerID1
	var i int
	for _, positionID := range []string{ProductID1SKU1, ProductID3SKU2} {
		i++
		positionVo := &Article{}
		positionVo.ID = positionID
		positionVo.Price = 100
		positionVo.Quantity = 1

		orderVo.Articles = append(orderVo.Articles, positionVo)

	}
	return orderVo, nil
}

func createMockOrderScaled(t *testing.T) (*ArticleCollection, error) {
	orderVo := &ArticleCollection{}
	orderVo.CustomerID = CustomerID1
	var i int
	for _, positionID := range []string{ProductID1SKU1, ProductID1SKU2, ProductID3SKU2} {
		i++
		positionVo := &Article{}

		positionVo.ID = positionID
		positionVo.Price = 100 * float64(i)
		positionVo.Quantity = float64(i * 2)

		orderVo.Articles = append(orderVo.Articles, positionVo)
	}
	return orderVo, nil
}

func createMockOrderXY(t *testing.T) (*ArticleCollection, error) {
	orderVo := &ArticleCollection{}

	orderVo.CustomerID = CustomerID1

	positionVo := &Article{}
	positionVo.ID = ProductID1SKU1
	positionVo.Price = 100
	positionVo.Quantity = 2
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID1SKU2
	positionVo.Price = 300
	positionVo.Quantity = float64(2)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	positionVo = &Article{}
	positionVo.ID = ProductID3SKU2
	positionVo.Price = 500
	positionVo.Quantity = float64(2)
	orderVo.Articles = append(orderVo.Articles, positionVo)

	return orderVo, nil
}

// func removeOrder(articleCollection *ArticleCollection) {
// 	articleCollection.Delete()
// }
