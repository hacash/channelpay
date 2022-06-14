package chanpay

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/event"
	"github.com/hacash/node/websocket"
	"sync"
	"sync/atomic"
	"time"
)

/**
 * 通道方
 */

// Channel connector
type ChannelSideConn struct {
	statusMux sync.RWMutex

	// Ws long connection
	WsConn *websocket.Conn

	// data
	ChannelId   fields.ChannelId             // Channel Chain ID
	ChannelInfo *protocol.RpcDataChannelInfo // Channel current information

	OurAddress    fields.Address // Our address
	RemoteAddress fields.Address // Opposite address (customer address or opposite address of settlement channel)

	// Latest reconciliation bill
	LatestReconciliationBalanceBill channel.ReconciliationBalanceBill

	// Latest heartbeat time
	lastestHeartbeatTime time.Time

	// Close collection mark
	businessCloseAutoCollectionStatus uint32

	// Payment collection status lock 0:未占用  1:占用状态
	businessExclusiveStatus uint32 //

	// Message subscription
	msgFeed     event.Feed
	msgFeedErrs []event.Subscription
}

func NewChannelSideById(cid fields.ChannelId) *ChannelSideConn {
	return &ChannelSideConn{
		WsConn:                            nil,
		ChannelId:                         cid,
		ChannelInfo:                       nil,
		LatestReconciliationBalanceBill:   nil,
		lastestHeartbeatTime:              time.Now(),
		businessExclusiveStatus:           0,
		businessCloseAutoCollectionStatus: 0,
		msgFeedErrs:                       make([]event.Subscription, 0),
	}
}
func NewChannelSideByConn(conn *websocket.Conn) *ChannelSideConn {
	return &ChannelSideConn{
		WsConn:                            conn,
		ChannelInfo:                       nil,
		LatestReconciliationBalanceBill:   nil,
		lastestHeartbeatTime:              time.Now(),
		businessExclusiveStatus:           0,
		businessCloseAutoCollectionStatus: 0,
		msgFeedErrs:                       make([]event.Subscription, 0),
	}
}

// Get heartbeat time

func (c *ChannelSideConn) GetLastestHeartbeatTime() time.Time {
	c.statusMux.RLock()
	defer c.statusMux.RUnlock()
	return c.lastestHeartbeatTime
}

// Start message listening
func (c *ChannelSideConn) StartMessageListen() {
	go func() {
		//defer fmt.Printf("ChannelSideConn %s message listen end.\n", c.WsConn.RemoteAddr())
		for {
			msg, _, e := protocol.ReceiveMsg(c.WsConn)
			if e != nil {
				for _, v := range c.msgFeedErrs {
					v.Unsubscribe() // Disconnect, auto unsubscribe
				}
				c.msgFeedErrs = make([]event.Subscription, 0) // empty
				return                                        // Error occurred, end
			}
			if msg.Type() == protocol.MsgTypeHeartbeat {
				c.statusMux.Lock()
				c.lastestHeartbeatTime = time.Now()
				c.statusMux.Unlock()
			}
			// Broadcast message
			c.msgFeed.Send(msg)
		}
	}()
}

// Subscription message processing
func (c *ChannelSideConn) SubscribeMessage(chanobj chan protocol.Message) event.Subscription {
	subobj := c.msgFeed.Subscribe(chanobj) // 订阅消息处理
	c.msgFeedErrs = append(c.msgFeedErrs, subobj)
	return subobj
}

func (c *ChannelSideConn) SetChannelId(id fields.ChannelId) {
	c.ChannelId = id
}

func (c *ChannelSideConn) GetChannelId() fields.ChannelId {
	return c.ChannelId
}

func (c *ChannelSideConn) SetChannelInfo(info *protocol.RpcDataChannelInfo) {
	c.ChannelInfo = info
}

func (c *ChannelSideConn) GetChannelInfo() *protocol.RpcDataChannelInfo {
	return c.ChannelInfo
}

func (c *ChannelSideConn) SetAddresses(our, remote fields.Address) {
	c.OurAddress = our
	c.RemoteAddress = remote
}

func (c *ChannelSideConn) GetOurAddress() fields.Address {
	return c.OurAddress
}

func (c *ChannelSideConn) GetRemoteAddress() fields.Address {
	return c.RemoteAddress
}

func (c *ChannelSideConn) SetReconciliationBill(bill channel.ReconciliationBalanceBill) {
	c.statusMux.Lock()
	defer c.statusMux.Unlock()

	c.LatestReconciliationBalanceBill = bill
}

func (c *ChannelSideConn) GetReconciliationBill() channel.ReconciliationBalanceBill {
	c.statusMux.RLock()
	defer c.statusMux.RUnlock()

	return c.LatestReconciliationBalanceBill
}

// Check whether the collection channel is occupied
func (c *ChannelSideConn) IsInBusinessExclusive() bool {
	// Check status
	return atomic.LoadUint32(&c.businessExclusiveStatus) == 1
}

// Enable state exclusive
func (c *ChannelSideConn) StartBusinessExclusive() bool {
	return atomic.CompareAndSwapUint32(&c.businessExclusiveStatus, 0, 1)
}

// Remove state exclusivity
func (c *ChannelSideConn) ClearBusinessExclusive() {
	atomic.CompareAndSwapUint32(&c.businessExclusiveStatus, 1, 0)
}

// Check whether automatic collection is closed
func (c *ChannelSideConn) IsInCloseAutoCollectionStatus() bool {
	// Check status
	return atomic.LoadUint32(&c.businessCloseAutoCollectionStatus) == 1
}

// 启用关闭自动收款
func (c *ChannelSideConn) StartCloseAutoCollectionStatus() bool {
	return atomic.CompareAndSwapUint32(&c.businessCloseAutoCollectionStatus, 0, 1)
}

// Cancel closing automatic collection
func (c *ChannelSideConn) ClearCloseAutoCollectionStatus() {
	atomic.CompareAndSwapUint32(&c.businessCloseAutoCollectionStatus, 1, 0)
}

// judge
func (c *ChannelSideConn) RemoteAddressIsLeft() bool {
	return c.RemoteAddress.Equal(c.ChannelInfo.LeftAddress)
}

// Get channel data
func (c *ChannelSideConn) GetAvailableReuseVersion() uint32 {
	c.statusMux.RLock()
	defer c.statusMux.RUnlock()

	var bill = c.LatestReconciliationBalanceBill
	if bill != nil {
		return bill.GetReuseVersion()
	}
	return uint32(c.ChannelInfo.ReuseVersion)
}
func (c *ChannelSideConn) GetAvailableAutoNumber() uint64 {
	c.statusMux.RLock()
	defer c.statusMux.RUnlock()

	var bill = c.LatestReconciliationBalanceBill
	if bill != nil {
		return bill.GetAutoNumber()
	}
	return uint64(0)
}

// Get channel capacity
// side = our, remote
func (c *ChannelSideConn) GetChannelCapacityAmount(side string) fields.Amount {
	c.statusMux.RLock()
	defer c.statusMux.RUnlock()

	leftAmt := c.ChannelInfo.LeftAmount
	rightAmt := c.ChannelInfo.RightAmount
	// Judge whether there is a receipt
	bill := c.LatestReconciliationBalanceBill
	if bill != nil {
		leftAmt = bill.GetLeftBalance()
		rightAmt = bill.GetRightBalance()
	}
	remoteIsLeft := c.RemoteAddress.Equal(c.ChannelInfo.LeftAddress)
	//fmt.Println(leftAmt.ToFinString(), rightAmt.ToFinString())
	// Return capacity
	if (side == "remote" && remoteIsLeft) ||
		(side == "our" && !remoteIsLeft) {
		return leftAmt
	} else {
		return rightAmt
	}

}
func (c *ChannelSideConn) GetChannelCapacityAmountOfOur() fields.Amount {
	return c.GetChannelCapacityAmount("our")
}
func (c *ChannelSideConn) GetChannelCapacityAmountOfRemote() fields.Amount {
	return c.GetChannelCapacityAmount("remote")
}

// Directly save (do not check) payment reconciliation bills
func (c *ChannelSideConn) UncheckSignSaveBillByCompleteDocuments(bills *channel.ChannelPayCompleteDocuments) error {

	// Find the statement
	var proveBody *channel.ChannelChainTransferProveBodyInfo = nil
	for _, v := range bills.ProveBodys.ProveBodys {
		if v.ChannelId.Equal(c.ChannelId) {
			proveBody = v
			break
		}
	}
	// Whether it exists
	if proveBody == nil {
		return fmt.Errorf("proveBody of channel id %s not find", c.ChannelId.ToHex())
	}
	// Check reconciliation serial number
	if c.ChannelInfo.ReuseVersion != proveBody.ReuseVersion {
		return fmt.Errorf("ReuseVersion not match need %d but got %d", c.ChannelInfo.ReuseVersion, proveBody.ReuseVersion)
	}
	needBillAutoNumber := fields.VarUint8(1)
	if c.LatestReconciliationBalanceBill != nil {
		needBillAutoNumber = fields.VarUint8(c.LatestReconciliationBalanceBill.GetAutoNumber() + 1)
	}
	if needBillAutoNumber != proveBody.BillAutoNumber {
		return fmt.Errorf("BillAutoNumber not match need %d but got %d", needBillAutoNumber, proveBody.BillAutoNumber)
	}

	// preservation
	c.SetReconciliationBill(&channel.OffChainCrossNodeSimplePaymentReconciliationBill{
		ChannelChainTransferTargetProveBody: *proveBody,
		ChannelChainTransferData:            *bills.ChainPayment,
	})

	// success
	return nil
}

/////////////////////////////////////////////////////

type ChannelSideConnWrap interface {
	GetChannelCapacityAmountForRemoteCollect() fields.Amount
}

// Sort by channel capacity
type ChannelSideConnListByCollectCapacity []ChannelSideConnWrap

func (c ChannelSideConnListByCollectCapacity) Len() int {
	return len(c)
}

func (n ChannelSideConnListByCollectCapacity) Less(i, j int) bool {
	//fmt.Println(i, j, n[i] < n[j], n)
	jamt := n[j].GetChannelCapacityAmountForRemoteCollect()
	if n[i].GetChannelCapacityAmountForRemoteCollect().LessThan(&jamt) {
		return true
	} else {
		return false
	}
}

func (n ChannelSideConnListByCollectCapacity) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
