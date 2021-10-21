package client

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
)

// 对账返回
func (c *ChannelPayClient) DealServicerRespondReconciliation(msg *protocol.MsgServicerRespondReconciliation) {

	c.user.changeMux.Lock()
	defer c.user.changeMux.Unlock()

	waitBill := c.user.waitRealtimeReconciliation
	if waitBill == nil {
		return // 没有单子
	}
	tarAddr := account.NewAddressFromPublicKeyV0(msg.SelfSign.PublicKey)
	if waitBill.LeftAddress.Equal(tarAddr) {
		waitBill.LeftSign = msg.SelfSign
	} else {
		waitBill.RightSign = msg.SelfSign
	}
	// 检查单据签名
	if e := waitBill.CheckAddressAndSign(); e != nil {
		c.user.waitRealtimeReconciliation = nil // 签名检查失败
		return
	}

	// 更新票据状态
	c.updateReconciliationBalanceBill(waitBill)
}

// 启动对账
func (c *ChannelPayClient) InitiateReconciliation(bill *channel.OffChainCrossNodeSimplePaymentReconciliationBill) {

	//c.changeMux.Lock()
	//defer c.changeMux.Unlock()

	// 状态
	chanside := c.user.servicerStreamSide.ChannelSide
	chanside.StartBusinessExclusive() // 状态独占
	defer chanside.ClearBusinessExclusive()

	// 转换为对账单
	recbill := bill.ConvertToRealtimeReconciliation()

	// 计算我方签名
	sign, _, e := recbill.FillTargetSignature(c.user.selfAcc)
	if e != nil {
		return // 失败
	}

	// 记录
	c.user.waitRealtimeReconciliation = recbill

	// 向服务端发送签名
	conn := chanside.WsConn
	e = protocol.SendMsg(conn, &protocol.MsgClientInitiateReconciliation{
		SelfSign: *sign,
	})
	if e != nil {
		fmt.Println(e.Error())
		return // 失败
	}
}
