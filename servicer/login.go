package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
)

// 初始化通道
func (s *Servicer) InitializeChannelSide(side *chanpay.ChannelSideConn, remoteAddress fields.Address, weIsLeft bool) error {
	if side == nil || side.ChannelId == nil {
		return fmt.Errorf("side or side.ChannelId is nil")
	}
	cid := side.ChannelId

	// 读取通道状态
	chanInfo, e := protocol.RequestRpcReqChannelInfo(s.config.FullNodeRpcUrl, cid)
	if e != nil {
		return fmt.Errorf("request channel info fail: %s", e.Error())
	}

	var remoteIsRight = true

	// 查询是否在channel服务列表内
	if remoteAddress != nil {
		// 客户服务通道
		serhav := s.chanset.CheckCustomerPayChannel(cid)
		if serhav == false {
			// 不在服务列表内
			return fmt.Errorf("channel %s is not in the service list.", cid.ToHex())
		}
		remoteIsLeft := remoteAddress.Equal(chanInfo.LeftAddress)
		remoteIsRight = remoteAddress.Equal(chanInfo.RightAddress)
		if !remoteIsLeft && !remoteIsRight {
			return fmt.Errorf("remoteAddress %s is not in the channel addresses.", remoteAddress.ToReadable())
		}
	} else {
		// 中继结算节点通道
		if weIsLeft {
			remoteAddress = chanInfo.RightAddress
			remoteIsRight = true
		} else {
			remoteAddress = chanInfo.LeftAddress
			remoteIsRight = false
		}
	}

	// 处理
	// 读取最新对账单
	bill, e := s.billstore.GetLastestBalanceBill(cid)
	if e != nil {
		return fmt.Errorf("load lastest balance bill error: %s", e.Error())
	}
	var chanLeftAddr = chanInfo.LeftAddress
	var chanRightAddr = chanInfo.RightAddress
	if bill != nil {
		side.SetReconciliationBill(bill)
		if chanLeftAddr.NotEqual(bill.GetLeftAddress()) ||
			chanRightAddr.NotEqual(bill.GetRightAddress()) {
			// 票据和通道地址不匹配
			return fmt.Errorf("channel and bill address not match.")
		}
	}

	// 通道和账户信息
	side.ChannelInfo = chanInfo // 通道Info
	if remoteIsRight {
		side.OurAddress = chanLeftAddr
		side.RemoteAddress = chanRightAddr
	} else {
		side.OurAddress = chanRightAddr
		side.RemoteAddress = chanLeftAddr
	}

	return nil
}

// 全新登录一个客户端
func (s *Servicer) LoginNewCustomer(newcur *chanpay.Customer) error {
	// 初始化
	e := s.InitializeChannelSide(newcur.ChannelSide, newcur.ChannelSide.RemoteAddress, false)
	if e != nil {
		return e
	}

	// 开始监听消息
	newcur.ChannelSide.StartMessageListen()

	// 登录成功
	return nil
}
