package servicer

import (
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/core/channel"
)

/**
 * 处理远程支付
 */

// Callback of successful payment
func (s *Servicer) callbackPaymentSuccessed(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill) {
	// passageway
	tarcid := newbill.GetChannelId()
	oldbill, e := s.billstore.GetLastestBalanceBill(tarcid)
	if e != nil {
		return // return
	}
	// preservation
	if oldbill == nil ||
		(newbill.GetReuseVersion() == oldbill.GetReuseVersion() &&
			newbill.GetAutoNumber() == oldbill.GetAutoNumber()+1) {
		// If the saving conditions are met, execute saving
		s.billstore.UpdateStoreBalanceBill(tarcid, newbill)
		//fmt.Println("UpdateStoreBalanceBill: ", tarcid.ToHex(), newbill.GetAutoNumber())
	}

	// Update channel side
	var side *chanpay.ChannelSideConn = nil
	// Search settlement channel
	s.settlenoderChgLock.RLock()
	for _, v := range s.settlenoder {
		for _, node := range v {
			if node.ChannelId.Equal(tarcid) {
				// eureka
				side = node.ChannelSide
				break
			}
		}
	}
	s.settlenoderChgLock.RUnlock()
	// Update ticket
	if side != nil {
		side.SetReconciliationBill(newbill)
	}

	// Query client connections
	s.customerChgLock.RLock()
	for _, u := range s.customers {
		if u.ChannelSide.ChannelId.Equal(tarcid) {
			// eureka
			side = u.ChannelSide
			break
		}
	}
	s.customerChgLock.RUnlock()
	// Update ticket
	if side != nil {
		side.SetReconciliationBill(newbill)
	}

	//lastbill := side.GetReconciliationBill()
	//fmt.Printf("1:%p, 2:%p, 3:%p, 4:%p\n", side, oldbill, newbill, lastbill)
	//
	//fmt.Println("callbackPaymentSuccessed: ", tarcid.ToHex(), newbill.GetAutoNumber(), lastbill.GetAutoNumber())
}
