package pricerule

import (
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/foomoShop/order"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------
const (
	ValidationPriceRuleMinimumAmount           TypeRuleValidationMsg = "minimum_order_amount_not_reached"            // if sum of ALL applicable prices*qtys < threshold
	ValidationPriceRuleMinimumApplicableAmount TypeRuleValidationMsg = "minimum_order_applicable_amount_not_reached" // if sum of applicable prices*qtys < threshold

	ValidationPriceRuleIncludeProductGroupsNotMatching  TypeRuleValidationMsg = "include_product_groups_not_matching"
	ValidationPriceRuleExcludeProductGroupsNotMatching  TypeRuleValidationMsg = "exclude_product_groups_not_matching"
	ValidationPriceRuleIncludeCustomerGroupsNotMatching TypeRuleValidationMsg = "include_customer_groups_not_matching"
	ValidationPriceRuleExcludeCustomerGroupsNotMatching TypeRuleValidationMsg = "exclude_customer_groups_not_matching"

	ValidationPriceRuleMaxUsages            TypeRuleValidationMsg = "max_usages_reached"
	ValidationPriceRuleMaxUsagesPerCustomer TypeRuleValidationMsg = "max_usages_per_customer_reached"

	ValidationPriceRuleExpired     TypeRuleValidationMsg = "pricerule_expired"
	ValidationPriceRuleNotValidYet TypeRuleValidationMsg = "pricerule_not_valid_yet"

	ValidationVoucherUnknown      TypeRuleValidationMsg = "voucher_unknown"
	ValidationVoucherPersonalized TypeRuleValidationMsg = "voucher_only_for_customer"
	ValidationVoucherAlreadyUsed  TypeRuleValidationMsg = "voucher_already_used"

	ValidationPreviouslyAppliedRuleBlock TypeRuleValidationMsg = "rule_blocked_by_previously_applied_rule" //rule application blocked
	ValidationPriceRuleOK                TypeRuleValidationMsg = "price_rule_ok"
)

//------------------------------------------------------------------
// ~ PUBLIC FUNCTIONS
//------------------------------------------------------------------

// ValidateVoucher - validates a voucher code and returns o = true or a validation message when ok = false
// if order is not provided it will only check non-order related conditions
// if customerID == "" we assume not-logged in or guest
//
// validationMessage is
//
// - ValidationPriceRuleOK - validation passed
//
// - ValidationVoucherUnknown - can not find voucher by code or voucher object corrupted
//
// - ValidationVoucherPersonalized - voucher personalized, but customer unknown (guest) or customer ids not matching
//
// - ValidationVoucherAlreadyUsed - PERSONALIZED voucher redeemed - (used on a placed/finalized order)
//
// - ValidationPriceRuleMaxUsages - priceRule.MaxUses <= priceRule.UsageHistory.TotalUsages
//
// - ValidationPriceRuleMaxUsagesPerCustomer - priceRule.MaxUsesPerCustomer <= customerUsages
//
// - ValidationPriceRuleMinimumAmount - minimum order amount not met
//
// - ValidationPriceRuleIncludeProductGroupsNotMatching - if no matching products on order - order can only be applied to certain product groups ...
//
// - ValidationPriceRuleExcludeProductGroupsNotMatching - if all the order items are included in the excluded product groups list - for example rule only for non-salwe items, but all items are sale ones
//
// - ValidationPriceRuleIncludeCustomerGroupsNotMatching - not the right customer group - for example rule only for employees
//
// - ValidationPriceRuleExcludeCustomerGroupsNotMatching - can not be applied for customers in the group ... for example rule not for employees
//
// - ValidationPreviouslyAppliedRuleBlock - a previously applied rule (with priority number higher) has a property set to true ... no further rules can be applied
func ValidateVoucher(voucherCode string, order *order.Order) (ok bool, validationMessage TypeRuleValidationMsg) {
	//check if voucher is for customer or generic/guest
	//get voucher

	customerID := order.CustomerData.CustomerId
	voucher, voucherPriceRule, err := GetVoucherAndPriceRule(voucherCode)

	//check if exists
	if err != nil || voucher.VoucherCode != voucherCode {
		return false, ValidationVoucherUnknown
	}

	//check if voucher personalized
	if voucher.VoucherType == VoucherTypePersonalized {
		//no customer info atm, guest or not logged in yet ...
		if len(customerID) == 0 && len(voucher.CustomerID) != 0 {
			return false, ValidationVoucherPersonalized
		}
	}

	//check if voucher was redeemed already
	if !voucher.TimeRedeemed.IsZero() && voucher.VoucherType == VoucherTypePersonalized {
		return false, ValidationVoucherAlreadyUsed
	}

	//--------------------------------------------------------------
	// PriceRule
	//--------------------------------------------------------------
	//find groups for customer
	groupIDsForCustomer := GetGroupsIDSForItem(order.CustomerData.CustomerId, CustomerGroup)
	if len(groupIDsForCustomer) == 0 {
		groupIDsForCustomer = []string{}
	}
	//find the groupIds for order items
	productGroupIDsPerPosition := getProductGroupIDsPerPosition(order)
	ok, priceRuleFailReason := validatePriceRuleForOrder(*voucherPriceRule, order, productGroupIDsPerPosition, groupIDsForCustomer)
	if !ok {
		return false, priceRuleFailReason
	}

	ok, priceRuleFailReason = checkPreviouslyAppliedRules(voucherPriceRule, voucher, order, groupIDsForCustomer, productGroupIDsPerPosition)
	if !ok {
		return false, priceRuleFailReason
	}
	return true, ValidationPriceRuleOK
}

// CommitDiscounts is called when and order is finalized - it redeems all personalized vouchers
// and updates the pricerule/voucher usage history
// IT IS IRREVERSIBLE!!!
//
// alternatively use CommitOrderDiscounts
func CommitDiscounts(orderDiscounts *OrderDiscounts, customerID string) error {
	var appliedRuleIDs []string
	var appliedVoucherRuleIDs []string
	var appliedVoucherCodes []string

	for _, orderDiscount := range *orderDiscounts {
		for _, appliedDiscount := range orderDiscount.AppliedDiscounts {
			//if voucher - we need to redeem vouchers so we keep the separately
			if len(appliedDiscount.VoucherCode) > 0 {
				appliedVoucherCodes = append(appliedVoucherCodes, appliedDiscount.VoucherCode)
				appliedVoucherRuleIDs = append(appliedVoucherRuleIDs, appliedDiscount.PriceRuleID)
			} else {
				//else normal rule
				appliedRuleIDs = append(appliedRuleIDs, appliedDiscount.PriceRuleID)
			}
		}
	}

	appliedRuleIDs = RemoveDuplicates(appliedRuleIDs)
	appliedVoucherRuleIDs = RemoveDuplicates(appliedVoucherRuleIDs)
	appliedVoucherCodes = RemoveDuplicates(appliedVoucherCodes)

	spew.Dump(appliedRuleIDs)
	spew.Dump(appliedVoucherRuleIDs)

	//redeem vouchers first
	//NOTE: redeem internaly manipulates the associated pricerule as well
	for _, voucherCode := range appliedVoucherCodes {

		err := redeemVoucherByCode(voucherCode, customerID)
		if err != nil {
			return err
		}
	}

	spew.Dump(appliedRuleIDs)
	for _, ruleID := range appliedRuleIDs {
		err := UpdatePriceRuleUsageHistoryAtomic(ruleID, customerID)
		if err != nil {
			return err
		}
	}
	return nil
}

// CommitOrderDiscounts is called when and order is finalized - it redeems all personalized vouchers
// and updates the pricerule/voucher usage history
// IT IS IRREVERSIBLE!!!
//
// alternatively use CommitDiscounts
func CommitOrderDiscounts(customerID string, order *order.Order, voucherCodes []string, paymentMethod string, roundTo float64) error {
	orderDiscounts, _, err := ApplyDiscounts(order, voucherCodes, paymentMethod, roundTo)
	if err != nil {
		return err
	}
	return CommitDiscounts(&orderDiscounts, customerID)
}

//------------------------------------------------------------------
// ~ PRIVATE FUNCTIONS
//------------------------------------------------------------------

// redeemVoucherByCode -
func redeemVoucherByCode(voucherCode string, customerID string) error {
	voucher, err := GetVoucherByCode(voucherCode, nil)
	if err != nil {
		return err
	}
	return voucher.Redeem(customerID)
}

// Returns false, ValidationPreviouslyAppliedRuleBlock if a previous rule blocks application
func checkPreviouslyAppliedRules(voucherPriceRule *PriceRule, voucher *Voucher, order *order.Order, groupIDsForCustomer []string, productGroupIDsPerPosition map[string][]string) (ok bool, reason TypeRuleValidationMsg) {
	// find applicable pricerules - auto promotions
	promotionPriceRules, err := GetValidPriceRulesForPromotions()

	if err != nil {
		panic(err)
	}

	var ruleVoucherPairs []RuleVoucherPair
	for _, promotionRule := range promotionPriceRules {
		ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: &promotionRule, Voucher: nil})
	}

	ruleVoucherPairs = append(ruleVoucherPairs, RuleVoucherPair{Rule: voucherPriceRule, Voucher: voucher})
	sort.Sort(ByPriority(ruleVoucherPairs))

	found := false
	for _, ruleVoucherPair := range ruleVoucherPairs {
		if voucher != nil && ruleVoucherPair.Voucher != nil {
			if ruleVoucherPair.Voucher.VoucherCode == voucher.VoucherCode {
				found = true
			}
			for _, position := range order.Positions {
				applicable, _ := validatePriceRuleForPosition(*ruleVoucherPair.Rule, order, position, productGroupIDsPerPosition, groupIDsForCustomer)
				if applicable && ruleVoucherPair.Rule.Exclusive == true {
					if found == false {
						return false, ValidationPreviouslyAppliedRuleBlock
					}
					return true, ValidationPreviouslyAppliedRuleBlock
				}
			}
		}
	}
	return true, ValidationPreviouslyAppliedRuleBlock
}
