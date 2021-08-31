package servicer

import (
	"fmt"
)

// 全新登录一个客户端
func (s *Servicer) LoginNewCustomer(newcur *Customer) error {

	// 读取通道状态
	chanInfo, e := s.rpcReqChannelInfo(newcur.channelId)
	if e != nil || chanInfo == nil {
		return fmt.Errorf("load channel info fail: %s", e.Error())
	}

	// 读取最新对账单
	bill, e := s.billstore.GetLastestBalanceBill(newcur.channelId)
	if e != nil {
		return e
	}
	var chanLeftAddr = chanInfo.LeftAddress
	var chanRightAddr = chanInfo.RightAddress
	if bill != nil {
		newcur.latestReconciliationBalanceBill = bill
		if chanLeftAddr.NotEqual(bill.LeftAddress()) ||
			chanRightAddr.NotEqual(bill.RightAddress()) {
			// 票据和通道地址不匹配
			return fmt.Errorf("channel and bill address not match.")
		}
	}

	// 账户
	cusAddr := newcur.customerAddress
	customerIsLeft := cusAddr.Equal(chanLeftAddr)
	customerIsRight := cusAddr.Equal(chanRightAddr)
	if !customerIsLeft && !customerIsRight {
		return fmt.Errorf("address %s is not belong to channel %s",
			cusAddr.ToReadable(),
			newcur.channelId.ToHex())
	}
	if customerIsLeft {
		newcur.servicerAddress = chanRightAddr
	} else {
		newcur.servicerAddress = chanLeftAddr
	}

	// 完成登录
	return nil
}
