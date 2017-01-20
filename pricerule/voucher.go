package pricerule

import (
	"errors"
	"log"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//------------------------------------------------------------------
// ~ CONSTANTS
//------------------------------------------------------------------
const (
	VoucherTypeAnonymous    VoucherType = "anonymous"    // guests and all
	VoucherTypePersonalized VoucherType = "personalized" // when customer is known
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

//Voucher - voucher vo
type Voucher struct {
	ID          string //voucher ID
	VoucherCode string
	PriceRuleID string // ID of the price rule
	MappingID   string
	VoucherType VoucherType //VoucherType.. personalized or anonymous
	CustomerID  string      // the customer if applicable

	TimeApplied  time.Time //used in cart but can still be used until redeemed
	TimeRedeemed time.Time //used on itemCollection - this is a redeem

	CreatedAt      time.Time //created at
	LastModifiedAt time.Time //updated at

	Custom interface{} `bson:",omitempty"` //make it extensible if needed

}

//VoucherType - voucher type
type VoucherType string

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewVoucher - sets defaults
func NewVoucher(ID string, voucherCode string, priceRule *PriceRule, customerID string) *Voucher {
	voucher := new(Voucher)
	voucher.ID = ID
	voucher.PriceRuleID = priceRule.ID
	voucher.MappingID = priceRule.MappingID
	voucher.VoucherCode = voucherCode
	if len(customerID) > 0 {
		voucher.VoucherType = VoucherTypePersonalized
	} else {
		voucher.VoucherType = VoucherTypeAnonymous
	}

	return voucher
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// LoadVoucher -
// make sure the error is then handled properly
func LoadVoucher(ID string, customProvider PriceRuleCustomProvider) (*Voucher, error) {
	return GetVoucherByID(ID, customProvider)
}

// VoucherAlreadyExistsInDB checks if a voucher with given ID already exists in the database
func VoucherAlreadyExistsInDB(ID string) (bool, error) {
	return ObjectOfTypeAlreadyExistsInDB(ID, new(Voucher))
}

// Upsert - upsers a Voucher
// note that if you programmatically manipulate the CreatedAt time, this methd will upsert it
func (voucher *Voucher) Upsert() error {
	//set created and modified times

	if voucher.CreatedAt.IsZero() {
		voucherFromDb, err := GetVoucherByID(voucher.ID, nil)
		if err != nil || voucherFromDb == nil {
			voucher.CreatedAt = time.Now()
		} else {
			voucher.CreatedAt = voucherFromDb.CreatedAt
		}

	}
	voucher.LastModifiedAt = time.Now()

	p := GetPersistorForObject(voucher)
	_, err := p.GetCollection().Upsert(bson.M{"id": voucher.ID}, voucher)

	if err != nil {
		return err
	}
	return nil
}

// SetApplied - set applied time and store
func (voucher *Voucher) SetApplied() error {
	voucher.TimeApplied = time.Now()
	return voucher.Upsert()
}

// Redeem - set redeem time and store
func (voucher *Voucher) Redeem(customerID string) error {
	mutex := sync.Mutex{}

	if voucher.VoucherType == VoucherTypePersonalized && len(voucher.CustomerID) > 0 {
		if voucher.CustomerID != customerID {
			return errors.New("voucher with ID " + voucher.ID + " not assigned to " + customerID + ". can not redeem")
		}
	}

	if !voucher.TimeRedeemed.IsZero() {
		return errors.New("voucher " + voucher.ID + " redeemed already")
	}

	mutex.Lock()
	defer mutex.Unlock()
	voucher.TimeRedeemed = time.Now()
	err := voucher.Upsert()
	log.Println("redeemed voucher " + voucher.VoucherCode)
	if err != nil {
		return err
	}
	return UpdatePriceRuleUsageHistoryAtomic(voucher.PriceRuleID, customerID)
}

// Delete - delete voucher - ID must be set
func (voucher *Voucher) Delete() error {
	err := GetPersistorForObject(voucher).GetCollection().Remove(bson.M{"id": voucher.ID})
	voucher = nil
	return err
}

// DeleteVoucher - delete voucher
func DeleteVoucher(ID string) error {
	err := GetPersistorForObject(new(Voucher)).GetCollection().Remove(bson.M{"id": ID})
	return err
}

// RemoveAllVouchers -
func RemoveAllVouchers() error {
	p := GetPersistorForObject(new(Voucher))
	_, err := p.GetCollection().RemoveAll(bson.M{})
	return err
}

// GetVoucherAndPriceRule -
func GetVoucherAndPriceRule(voucherCode string) (*Voucher, *PriceRule, error) {
	voucher, err := GetVoucherByCode(voucherCode, nil)
	if err != nil {
		return nil, nil, err
	}

	if voucher != nil && len(voucher.PriceRuleID) > 0 {
		//get the pricerule
		priceRule, err := GetPriceRuleByID(voucher.PriceRuleID, nil)
		if err != nil {
			return voucher, nil, err
		}
		return voucher, priceRule, nil
	}
	return nil, nil, errors.New("voucher with code " + voucherCode + " not found")
}
