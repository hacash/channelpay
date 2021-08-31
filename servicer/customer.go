package servicer

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
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
}

func NewCustomer(ws *websocket.Conn) *Customer {
	return &Customer{
		IsRegistered: false,
		wsConn:       ws,
		channelId:    nil,
		channelInfo:  nil,
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
