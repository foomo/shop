package pricerule

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/foomo/shop/configuration"
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

func init() {
	fmt.Println("Initializing persistors...")
	configuration.MONGO_URL = "mongodb://" + configuration.LocalUnitTests
}

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
func TestScaled(t *testing.T) {
	//Init
	RemoveAllGroups()
	RemoveAllPriceRules()
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
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionScaled

	priceRule.ScaledAmounts = append(priceRule.ScaledAmounts, ScaledAmountLevel{FromValue: 100.0, ToValue: 150.0, Amount: 10, IsScaledAmountPercentage: true})
	priceRule.ScaledAmounts = append(priceRule.ScaledAmounts, ScaledAmountLevel{FromValue: 150.01, ToValue: 100000.0, Amount: 50, IsScaledAmountPercentage: true})

	priceRule.MaxUses = 10
	priceRule.MaxUsesPerCustomer = 10
	priceRule.WhichXYFree = XYCheapestFree
	priceRule.IncludedProductGroupIDS = []string{groupID}
	priceRule.IncludedCustomerGroupIDS = []string{}

	//insert as well
	group.AddGroupItemIDs([]string{ProductID1SKU1})

	err := group.Upsert()
	if err != nil {
		log.Println(err)
	}

	//create articleCollection
	orderVo, err := createMockOrderScaled(t)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{""}, "", 0.05)
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
func TestBuyXGetY(t *testing.T) {
	//Init
	RemoveAllGroups()
	RemoveAllPriceRules()

	//create pricerule
	priceRule := NewPriceRule(PriceRuleIDSale)
	priceRule.Name = map[string]string{
		"de": PriceRuleIDSale,
		"fr": PriceRuleIDSale,
		"it": PriceRuleIDSale,
	}
	priceRule.Type = TypePromotionOrder
	priceRule.Description = priceRule.Name
	priceRule.Action = ActionBuyXGetY
	priceRule.X = 3
	priceRule.Y = 1
	priceRule.MaxUses = 10
	priceRule.MaxUsesPerCustomer = 10
	priceRule.IncludedProductGroupIDS = []string{}
	priceRule.IncludedCustomerGroupIDS = []string{}

	err := priceRule.Upsert()
	if err != nil {
		panic(err)
	}

	//create articleCollection
	orderVo, err := createMockOrderXY(t)
	if err != nil {
		panic(err)
	}

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{""}, "", 0.05)
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	fmt.Println("discounts for buy x get y")
	spew.Dump(discountsVo)
	spew.Dump(*summary)

}

// Test groups creation
func TestExclude(t *testing.T) {
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

	productGroupIDsPerPosition := getProductGroupIDsPerPosition(orderVo)
	//spew.Dump(productGroupIDsPerPosition)
	for _, article := range orderVo.Articles {
		ok, _ := validatePriceRuleForPosition(*priceRule, orderVo, article, productGroupIDsPerPosition, []string{})
		log.Println(article.ID + " " + priceRule.ID + " " + strconv.FormatBool(ok))
	}

	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, "blah", 0.05)
	spew.Dump(discountsVo)
	spew.Dump(*summary)

}

// Test groups creation
func TestMaxOrder(t *testing.T) {
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
	priceRule.MinOrderAmount = 100
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
	priceRule.MinOrderAmount = 1000
	priceRule.MinOrderAmountApplicableItemsOnly = false
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

	// PRICERULES --------------------------------------------------------------------------------------
	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{}, PaymentMethodID1, 0.05)
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
func TestTwoStepWorkflow(t *testing.T) {
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
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode2, VoucherCode1}, PaymentMethodID1, 0.05)
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
func TestPricerulesWorkflow(t *testing.T) {
	//remove all and add again
	Init(t)

	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}
	now := time.Now()
	discountsVo, summary, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode1}, PaymentMethodID1, 0.05)
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
func TestCheckoutWorkflow(t *testing.T) {
	//remove all and add again
	Init(t)

	orderVo, err := createMockOrder(t)
	if err != nil {
		panic(err)
	}
	now := time.Now()
	discountsVo, _, err := ApplyDiscounts(orderVo, nil, []string{VoucherCode1}, PaymentMethodID1, 0.05)
	timeTrack(now, "Apply multiple price rules")
	// defer removeOrder(orderVo)
	if err != nil {
		panic(err)
	}

	now = time.Now()
	ok, reason := ValidateVoucher(VoucherCode1, orderVo)
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

	priceRule.IncludedPaymentMethods = []string{PaymentMethodID1}

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
	var i int
	for _, positionID := range []string{ProductID1SKU1, ProductID1SKU2, ProductID3SKU2} {
		i++
		positionVo := &Article{}

		positionVo.ID = positionID
		positionVo.Price = 100 * float64(i)
		positionVo.Quantity = float64(1)

		orderVo.Articles = append(orderVo.Articles, positionVo)
	}
	return orderVo, nil
}

// func removeOrder(articleCollection *ArticleCollection) {
// 	articleCollection.Delete()
// }
