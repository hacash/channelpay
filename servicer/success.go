package servicer

import (
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/core/channel"
)

/**
 * 处理远程支付
 */

// 支付成功的回调
func (s *Servicer) callbackPaymentSuccessed(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill) {
	// 通道
	tarcid := newbill.GetChannelId()
	oldbill, e := s.billstore.GetLastestBalanceBill(tarcid)
	if e != nil {
		return // 返回
	}
	// 保存
	if oldbill == nil ||
		(newbill.GetReuseVersion() == oldbill.GetReuseVersion() &&
			newbill.GetAutoNumber() == oldbill.GetAutoNumber()+1) {
		// 符合保存条件，执行保存
		s.billstore.UpdateStoreBalanceBill(tarcid, newbill)
		//fmt.Println("UpdateStoreBalanceBill: ", tarcid.ToHex(), newbill.GetAutoNumber())
	}

	// 更新通道侧
	var side *chanpay.ChannelSideConn = nil
	// 搜索结算通道
	s.settlenoderChgLock.RLock()
	for _, v := range s.settlenoder {
		for _, node := range v {
			if node.ChannelId.Equal(tarcid) {
				// 找到了
				side = node.ChannelSide
				break
			}
		}
	}
	s.settlenoderChgLock.RUnlock()
	// 更新票据
	if side != nil {
		side.SetReconciliationBill(newbill)
	}

	// 查询客户端连接
	s.customerChgLock.RLock()
	for _, u := range s.customers {
		if u.ChannelSide.ChannelId.Equal(tarcid) {
			// 找到了
			side = u.ChannelSide
			break
		}
	}
	s.customerChgLock.RUnlock()
	// 更新票据
	if side != nil {
		side.SetReconciliationBill(newbill)
	}

	//lastbill := side.GetReconciliationBill()
	//fmt.Printf("1:%p, 2:%p, 3:%p, 4:%p\n", side, oldbill, newbill, lastbill)
	//
	//fmt.Println("callbackPaymentSuccessed: ", tarcid.ToHex(), newbill.GetAutoNumber(), lastbill.GetAutoNumber())
}
