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

	RegisteredID uint64 // Whether the registration has been completed and a random number will be assigned when it is completed

	LanguageSet fields.StringMax255 // Language settings en_ US zh_ CN

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

// Update heartbeat time
func (c *Customer) UpdateLastestHeartbeatTime() {
	c.updateMux.Lock()
	defer c.updateMux.Unlock()
	//fmt.Println("c *Customer UpdateLastestHeartbeatTime ", time.Now().Unix())
	c.lastestHeartbeatTime = time.Now()
	// Resume heartbeat
	protocol.SendMsg(c.ChannelSide.WsConn, &protocol.MsgHeartbeat{})
}
func (c *Customer) GetLastestHeartbeatTime() time.Time {
	c.updateMux.RLock()
	defer c.updateMux.RUnlock()
	return c.lastestHeartbeatTime
}

// Perform registration
func (c *Customer) DoRegister(channelId fields.ChannelId, address fields.Address) {
	c.RegisteredID = rand.Uint64()
	c.ChannelSide.ChannelId = channelId
	c.ChannelSide.RemoteAddress = address
}

// Replaced offline
func (c *Customer) DoDisplacementOffline(newcur *Customer) {
	// Copy data
	newcur.ChannelSide.LatestReconciliationBalanceBill = c.ChannelSide.LatestReconciliationBalanceBill
	newcur.ChannelSide.RemoteAddress = c.ChannelSide.RemoteAddress
	// Send the replaced message, and the replaced will be automatically offline
	protocol.SendMsg(c.ChannelSide.WsConn, &protocol.MsgDisplacementOffline{})
	// Close connection
	//fmt.Println("protocol.SendMsg(c.ChannelSide.WsConn, &protocol.MsgDisplacementOffline{})", c.ChannelSide.WsConn.RemoteAddr())
	c.ChannelSide.WsConn.Close()
}

// Check whether the collection channel is occupied
func (c *Customer) IsInBusinessExclusive() bool {
	return c.ChannelSide.IsInBusinessExclusive()
}

// Where state exclusive
func (c *Customer) StartBusinessExclusive() bool {
	return c.ChannelSide.StartBusinessExclusive()
}

// Remove state exclusivity
func (c *Customer) ClearBusinessExclusive() {
	c.ChannelSide.ClearBusinessExclusive()
}

// judge
func (c *Customer) GetCustomerAddress() fields.Address {
	return c.ChannelSide.RemoteAddress
}
func (c *Customer) GetServicerAddress() fields.Address {
	return c.ChannelSide.OurAddress
}

// judge
func (c *Customer) CustomerAddressIsLeft() bool {
	return c.ChannelSide.RemoteAddressIsLeft()
}

func (c *Customer) GetChannelCapacityAmountForRemotePay() fields.Amount {
	return c.ChannelSide.GetChannelCapacityAmountOfRemote()
}

func (c *Customer) GetChannelCapacityAmountForRemoteCollect() fields.Amount {
	return c.ChannelSide.GetChannelCapacityAmountOfOur()
}
