package client

import (
	"fmt"
	"github.com/hacash/core/channel"
)

// 支付成功的回调
func (c *ChannelPayClient) callbackPaymentSuccessed(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	showError := func(err string) {
		c.ShowLogString(err, false, true)
	}

	// 通道对不上
	if false == newbill.GetChannelId().Equal(c.user.chanInfo.ChannelId) {
		return // 忽略错误并返回
	}

	// 检查票据签名
	if nil != newbill.ChannelChainTransferData.CheckMustAddressAndSigns() {
		return // 忽略错误并返回
	}

	// 检查票据流水号
	lcbill := c.user.localLatestReconciliationBalanceBill
	lcruv := uint32(1)
	lcatn := uint64(0)
	if lcbill != nil {
		lcruv, lcatn = lcbill.GetReuseVersionAndAutoNumber()
	}

	// 检查票据状态
	newruv, newatn := newbill.GetReuseVersionAndAutoNumber()
	if lcruv != newruv {
		return // 忽略错误，直接返回
	}
	if newatn != lcatn+1 {
		showError(fmt.Sprintf("Callback payment successed error: %d != %d + 1",
			newatn, lcatn))
		return // 流水号对不上，错误返回
	}

	// 更新票据状态
	c.updateReconciliationBalanceBill(newbill)

	// 发起对账
	go c.InitiateReconciliation(newbill)
}

// 修改票据状态
func (c *ChannelPayClient) updateReconciliationBalanceBill(newbill channel.ReconciliationBalanceBill) {

	// 修改票据和余额状态
	var servicerSide = c.user.servicerStreamSide.ChannelSide
	servicerSide.SetReconciliationBill(newbill)

	// 保存票据至磁盘
	e := c.user.SaveLastBillToDisk(newbill)
	if e != nil {
		return // 出错,返回
	}
	//fmt.Println("ChannelPayClient SetReconciliationBill: ", newbill.GetAutoNumber())

	// 重新显示余额
	go c.UpdateBalanceShow()
}
