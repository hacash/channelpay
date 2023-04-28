package client

import (
	"fmt"
	"github.com/hacash/core/channel"
)

// Callback of successful payment
func (c *ChannelPayClient) callbackPaymentSuccessed(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	showError := func(err string) {
		c.ShowLogString(err, false, true)
	}

	// Channel mismatch
	if false == newbill.GetChannelId().Equal(c.user.chanInfo.ChannelId) {
		return // Ignore errors and return
	}

	// Check bill signature
	if nil != newbill.ChannelChainTransferData.CheckMustAddressAndSigns() {
		return // Ignore errors and return
	}

	// Check bill serial number
	lcbill := c.user.localLatestReconciliationBalanceBill
	lcruv := uint32(1)
	lcatn := uint64(0)
	if lcbill != nil {
		lcruv, lcatn = lcbill.GetReuseVersionAndAutoNumber()
	}

	// Check bill status
	newruv, newatn := newbill.GetReuseVersionAndAutoNumber()
	if lcruv != newruv {
		return // Ignore the error and return directly
	}
	if newatn != lcatn+1 {
		showError(fmt.Sprintf("Callback payment successed error: %d != %d + 1",
			newatn, lcatn))
		return // Serial number cannot be matched, error returned
	}

	// Update bill status
	c.updateReconciliationBalanceBill(newbill)

	// Initiate reconciliation
	go c.InitiateReconciliation(newbill)
}

// Modify bill status
func (c *ChannelPayClient) updateReconciliationBalanceBill(newbill channel.ReconciliationBalanceBill) {

	// Modify bill and balance status
	var servicerSide = c.user.servicerStreamSide.ChannelSide
	servicerSide.SetReconciliationBill(newbill)

	// Save ticket to disk
	e := c.user.SaveLastBillToDisk(newbill)
	if e != nil {
		return // Error, return
	}
	// fmt.Println("ChannelPayClient SetReconciliationBill: ", newbill.TypeCode(), newbill.GetAutoNumber())

	// Redisplay balance
	go c.UpdateBalanceShow()
}
