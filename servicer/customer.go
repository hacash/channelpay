package servicer

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
)

type Customer struct {
	IsRegistered bool // 是否已完成注册

	LanguageSet fields.StringMax255 // 语言设置 en_US zh_CN

	ChannelSide *ChannelSideConn
	//
	//
	//// 客户端长连接
	//wsConn *websocket.Conn
	//
	//// 数据
	//channelId   fields.Bytes16      // 通道链 ID
	//channelInfo *RpcDataChannelInfo // 通道当前的信息
	//
	//servicerAddress fields.Address // 服务节点地址
	//customerAddress fields.Address // 客户地址
	//
	//// 最新的对账票据
	//latestReconciliationBalanceBill channel.ReconciliationBalanceBill
	//
	//// 支付收款状态锁 0:未占用  1:占用状态
	//payBusinessExclusiveStatus uint32 //

}

func NewCustomer(ws *websocket.Conn) *Customer {
	side := NewChannelSideConn(ws)
	return &Customer{
		IsRegistered: false,
		ChannelSide:  side,
	}
}

func CreateChannelSideConnWrapForCustomer(list []*Customer) ChannelSideConnListByCollectCapacity {
	var res = make([]ChannelSideConnWrap, len(list))
	for i, v := range list {
		res[i] = v
	}
	return res
}

// 执行注册
func (c *Customer) DoRegister(channelId fields.Bytes16, address fields.Address) {
	c.IsRegistered = true
	c.ChannelSide.channelId = channelId
	c.ChannelSide.remoteAddress = address
}

// 被顶替下线
func (c *Customer) DoDisplacementOffline(newcur *Customer) {
	// 拷贝数据
	newcur.ChannelSide.latestReconciliationBalanceBill = c.ChannelSide.latestReconciliationBalanceBill
	newcur.ChannelSide.remoteAddress = c.ChannelSide.remoteAddress
	// 发送被顶替消息，被顶替者自动下线
	protocol.SendMsg(c.ChannelSide.wsConn, &protocol.MsgDisplacementOffline{})
	// 关闭连接
	c.ChannelSide.wsConn.Close()
}

// 检查收款通道是否被占用
func (c *Customer) IsInBusinessExclusive() bool {
	return c.ChannelSide.IsInBusinessExclusive()
}

// 其中状态独占
func (c *Customer) StartBusinessExclusive() bool {
	return c.ChannelSide.StartBusinessExclusive()
}

// 解除状态独占
func (c *Customer) ClearBusinessExclusive() {
	c.ChannelSide.ClearBusinessExclusive()
}

// 判断
func (c *Customer) GetCustomerAddress() fields.Address {
	return c.ChannelSide.remoteAddress
}
func (c *Customer) GetServicerAddress() fields.Address {
	return c.ChannelSide.ourAddress
}

// 判断
func (c *Customer) CustomerAddressIsLeft() bool {
	return c.ChannelSide.RemoteAddressIsLeft()
}

func (c *Customer) GetChannelCapacityAmountForRemotePay() fields.Amount {
	return c.ChannelSide.GetChannelCapacityAmountOfRemote()
}

func (c *Customer) GetChannelCapacityAmountForRemoteCollect() fields.Amount {
	return c.ChannelSide.GetChannelCapacityAmountOfOur()
}
