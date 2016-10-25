package pricerule

func createPriceRule() {

	CreateProductGroup()
	priceRule := NewPriceRule("ruleID")
	priceRule.Name = map[string]string{
		"de": "Rulename",
		"fr": "Rulename",
		"it": "Rulename",
	}

	priceRule.Description = map[string]string{
		"de": "Description",
		"fr": "Description",
		"it": "Description",
	}
	priceRule.Type = TypePromotion

	priceRule.Action = ActionItemByAbsolute

	priceRule.Amount = 10.0

	priceRule.Priority = 10
	priceRule.ActionType = "402"
	priceRule.PromoID = "5273639"
	priceRule.IncludedProductGroupIDS = []string{"SaleProducts"}

	// priceRule.IncludedCustomerGroupIDS = []string{CustomerGroupID1}

	err := priceRule.Upsert()
	if err != nil {
		panic(err)
	}
}

func CreateProductGroup() {
	group := new(Group)
	group.Type = ProductGroup
	group.ID = "SaleProducts"
	group.Name = "SaleProducts"
	group.AddGroupItemIDs([]string{"12387678"})
	err := group.Upsert()
	if err != nil {
		panic(err)
	}

}
