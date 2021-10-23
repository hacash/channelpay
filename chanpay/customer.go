package chanpay

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"math/rand"
	"sync"
	"time"
)

type Customer struct {
	updateMux sync.RWMutex

	RegisteredID uint64 // 是否已完成注册，完成时分配一个随机编号

	LanguageSet fields.StringMax255 // 语言设置 en_US zh_CN

	ChannelSide *ChannelSideConn

	lastestHeartbeatTime time.Time

	//
	//
	//// 客户端长连接
	//wsConn *websocket.Conn
	//
	//// 数据
	//channelId   fields.ChannelId      // 通道链 ID
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
	side := NewChannelSideByConn(ws)
	return &Customer{
		RegisteredID:         0,
		ChannelSide:          side,
		lastestHeartbeatTime: time.Now(),
	}
}

func CreateChannelSideConnWrapForCustomer(list []*Customer) ChannelSideConnListByCollectCapacity {
	var res = make([]ChannelSideConnWrap, len(list))
	for i, v := range list {
		res[i] = v
	}
	return res
}

// 更新心跳时间
func (c *Customer) UpdateLastestHeartbeatTime() {
	c.updateMux.Lock()
	defer c.updateMux.Unlock()
	c.lastestHeartbeatTime = time.Now()
}
func (c *Customer) GetLastestHeartbeatTime() time.Time {
	c.updateMux.RLock()
	defer c.updateMux.RUnlock()
	return c.lastestHeartbeatTime
}

// 执行注册
func (c *Customer) DoRegister(channelId fields.ChannelId, address fields.Address) {
	c.RegisteredID = rand.Uint64()
	c.ChannelSide.ChannelId = channelId
	c.ChannelSide.RemoteAddress = address
}

// 被顶替下线
func (c *Customer) DoDisplacementOffline(newcur *Customer) {
	// 拷贝数据
	newcur.ChannelSide.LatestReconciliationBalanceBill = c.ChannelSide.LatestReconciliationBalanceBill
	newcur.ChannelSide.RemoteAddress = c.ChannelSide.RemoteAddress
	// 发送被顶替消息，被顶替者自动下线
	protocol.SendMsg(c.ChannelSide.WsConn, &protocol.MsgDisplacementOffline{})
	// 关闭连接
	//fmt.Println("protocol.SendMsg(c.ChannelSide.WsConn, &protocol.MsgDisplacementOffline{})", c.ChannelSide.WsConn.RemoteAddr())
	c.ChannelSide.WsConn.Close()
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
	return c.ChannelSide.RemoteAddress
}
func (c *Customer) GetServicerAddress() fields.Address {
	return c.ChannelSide.OurAddress
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
