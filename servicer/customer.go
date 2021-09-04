package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"sync/atomic"
)

type Customer struct {
	IsRegistered bool // 是否已完成注册

	LanguageSet fields.StringMax255 // 语言设置 en_US zh_CN

	// 客户端长连接
	wsConn *websocket.Conn

	// 数据
	channelId   fields.Bytes16      // 通道链 ID
	channelInfo *RpcDataChannelInfo // 通道当前的信息

	servicerAddress fields.Address // 服务节点地址
	customerAddress fields.Address // 客户地址

	// 最新的对账票据
	latestReconciliationBalanceBill channel.ReconciliationBalanceBill

	// 支付收款状态锁 0:未占用  1:占用状态
	payBusinessExclusiveStatus uint32 //

}

func NewCustomer(ws *websocket.Conn) *Customer {
	return &Customer{
		IsRegistered:               false,
		wsConn:                     ws,
		channelId:                  nil,
		channelInfo:                nil,
		payBusinessExclusiveStatus: 0,
	}
}

// 执行注册
func (c *Customer) DoRegister(channelId fields.Bytes16, address fields.Address) {
	c.IsRegistered = true
	c.channelId = channelId
	c.customerAddress = address
}

// 被顶替下线
func (c *Customer) DoDisplacementOffline(newcur *Customer) {
	// 拷贝数据
	newcur.latestReconciliationBalanceBill = c.latestReconciliationBalanceBill
	newcur.servicerAddress = c.servicerAddress
	// 发送被顶替消息，被顶替者自动下线
	protocol.SendMsg(c.wsConn, &protocol.MsgDisplacementOffline{})
	// 关闭连接
	c.wsConn.Close()
}

// 检查收款通道是否被占用
func (c *Customer) IsInBusinessExclusive() bool {
	// 检查状态
	return atomic.LoadUint32(&c.payBusinessExclusiveStatus) == 1
}

// 其中状态独占
func (c *Customer) StartBusinessExclusive() bool {
	return atomic.CompareAndSwapUint32(&c.payBusinessExclusiveStatus, 0, 1)
}

// 解除状态独占
func (c *Customer) ClearBusinessExclusive() {
	atomic.CompareAndSwapUint32(&c.payBusinessExclusiveStatus, 1, 0)
}

// 判断
func (c *Customer) CustomerAddressIsLeft() bool {
	return c.customerAddress.Equal(c.channelInfo.LeftAddress)
}

// 获取通道容量
// kind = pay, collect
func (c *Customer) GetChannelCapacityAmount(kind string) *fields.Amount {
	leftAmt := c.channelInfo.LeftAmount
	rightAmt := c.channelInfo.RightAmount
	// 判断是否有收据
	bill := c.latestReconciliationBalanceBill
	if bill != nil {
		leftAmt = bill.LeftAmount()
		rightAmt = bill.RightAmount()
	}
	leftIsCustomer := c.customerAddress.Equal(c.channelInfo.LeftAddress)
	// 返回容量
	if kind == "pay" {
		if leftIsCustomer {
			return &leftAmt
		} else {
			return &rightAmt
		}
	} else if kind == "collect" {
		if leftIsCustomer {
			return &rightAmt
		} else {
			return &leftAmt
		}
	} else {
		emt := fields.NewEmptyAmount()
		return emt
	}
}
func (c *Customer) GetChannelCapacityAmountForPay() *fields.Amount {
	return c.GetChannelCapacityAmount("pay")
}
func (c *Customer) GetChannelCapacityAmountForCollect() *fields.Amount {
	return c.GetChannelCapacityAmount("collect")
}

// 直接保存（不做检查）支付对账票据
func (c *Customer) UncheckSignSaveChannelPayReconciliationBalanceBill(bills *channel.ChannelPayCompleteDocuments) error {

	// 找出对账单
	var proveBody *channel.ChannelChainTransferProveBodyInfo = nil
	for _, v := range bills.ProveBodys.ProveBodys {
		if v.ChannelId.Equal(c.channelId) {
			proveBody = v
			break
		}
	}
	// 是否存在
	if proveBody == nil {
		return fmt.Errorf("proveBody of channel id %s not find", c.channelId.ToHex())
	}
	// 检查对账流水号
	if c.channelInfo.ReuseVersion != proveBody.ChannelReuseVersion {
		return fmt.Errorf("ReuseVersion not match need %d but got %d", c.channelInfo.ReuseVersion, proveBody.ChannelReuseVersion)
	}
	needBillAutoNumber := fields.VarUint8(1)
	if c.latestReconciliationBalanceBill != nil {
		needBillAutoNumber = fields.VarUint8(c.latestReconciliationBalanceBill.ChannelAutoNumber() + 1)
	}
	if needBillAutoNumber != proveBody.BillAutoNumber {
		return fmt.Errorf("BillAutoNumber not match need %d but got %d", needBillAutoNumber, proveBody.BillAutoNumber)
	}

	// 保存
	c.latestReconciliationBalanceBill = &channel.CrossNodeSimplePaymentReconciliationBill{
		LeftAddr:                            c.channelInfo.LeftAddress,
		RightAddr:                           c.channelInfo.RightAddress,
		ChannelChainTransferTargetProveBody: *proveBody,
		ChannelChainTransferData:            *bills.ChainPayment,
	}

	// 成功
	return nil
}
