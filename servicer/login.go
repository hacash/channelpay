package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
)

// 全新登录一个客户端
func (s *Servicer) LoginNewCustomer(newcur *chanpay.Customer) error {

	cid := newcur.ChannelSide.GetChannelId()

	// 查询是否在channel服务列表内
	serhav := s.chanset.CheckCustomerPayChannel(cid)
	if serhav == false {
		// 不在服务列表内
		return fmt.Errorf("channel %s is not in the service list.", cid.ToHex())
	}

	// 读取通道状态
	chanInfo, e := protocol.RequestRpcReqChannelInfo(s.config.FullNodeRpcUrl, newcur.ChannelSide.ChannelId)
	if e != nil || chanInfo == nil {
		return fmt.Errorf("request channel info fail: %s", e.Error())
	}

	// 读取最新对账单
	bill, e := s.billstore.GetLastestBalanceBill(cid)
	if e != nil {
		return e
	}
	var chanLeftAddr = chanInfo.LeftAddress
	var chanRightAddr = chanInfo.RightAddress
	if bill != nil {
		newcur.ChannelSide.SetReconciliationBill(bill)
		if chanLeftAddr.NotEqual(bill.GetLeftAddress()) ||
			chanRightAddr.NotEqual(bill.GetRightAddress()) {
			// 票据和通道地址不匹配
			return fmt.Errorf("channel and bill address not match.")
		}
	}

	// 账户
	cusAddr := newcur.ChannelSide.RemoteAddress
	customerIsLeft := cusAddr.Equal(chanLeftAddr)
	customerIsRight := cusAddr.Equal(chanRightAddr)
	if !customerIsLeft && !customerIsRight {
		return fmt.Errorf("address %s is not belong to channel %s",
			cusAddr.ToReadable(),
			newcur.ChannelSide.ChannelId.ToHex())
	}
	if customerIsLeft {
		newcur.ChannelSide.OurAddress = chanRightAddr
	} else {
		newcur.ChannelSide.OurAddress = chanLeftAddr
	}

	// 完成登录
	return nil
}
