package chanpay

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/mint/event"
	"github.com/hacash/node/websocket"
	"strings"
	"sync"
	"time"
)

// journal
type PayActionLog struct {
	IsEnd     bool
	IsSuccess bool
	IsError   bool
	Content   string
}

// Instance type
type PayActionInstanceType uint8

const (
	PaySide     PayActionInstanceType = 1
	RelayNode   PayActionInstanceType = 2
	CollectSide PayActionInstanceType = 3
)

/**
* 单次支付行为操作
* 初始化调用：
0. SetUpstreamSide or ...
1. SetMustSignAddresses
2. StartOneSideMessageSubscription
3. InitCreateEmptyBillDocumentsByInitPayMsg
*
*/
type ChannelPayActionInstance struct {
	isBeDestroyed bool // Has been destroyed

	// Status lock
	statusUpdateMux sync.Mutex

	// Local service node
	localServicerNode *payroutes.PayRelayNode

	// Upstream and downstream connection ends of channel chain
	upstreamSide   *RelayPaySettleNoder // Channel upstream
	downstreamSide *RelayPaySettleNoder // Downstream of channel

	// Or the payer or the payee
	payCustomer     *Customer // Payment end client
	collectCustomer *Customer // Payee client

	// Channel length
	channelLength                    int  // Channel path length
	ourProveBodyIndex                int  // My channel path arrangement position
	ourProveBodyCompleted            bool // Have we completed the signature ourselves
	downstreamSubmitProveBodyCheckOK bool // Check the downstream submitted statement OK

	// Transaction start-up information
	payInitMsg *protocol.MsgRequestInitiatePayment //

	// Target transaction note
	billDocuments            *channel.ChannelPayCompleteDocuments
	transactionDistinguishId fields.VarUint8
	allSignsCompleted        bool

	// Signing machine
	signMachine DataSourceOfSignatureMachine

	// Signature address
	ourMustSignAddresses []fields.Address
	ourSignCompleted     bool // Have we completed the signature ourselves

	// Message subscriber
	msgFeedObjs []event.Subscription
	// Log subscriber
	logFeeds    event.Feed
	logFeedObjs []event.Subscription

	// Callback of successful payment
	successedBackCall []func(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill)

	// timeout handler 
	clearTimeout chan bool
}

// Newly created
func NewChannelPayActionInstance() *ChannelPayActionInstance {
	// establish
	ins := &ChannelPayActionInstance{
		isBeDestroyed:                    false,
		channelLength:                    0,
		ourProveBodyIndex:                0,
		localServicerNode:                nil,
		allSignsCompleted:                false,
		transactionDistinguishId:         0,
		signMachine:                      nil,
		ourMustSignAddresses:             nil, // 必须签名地址
		ourSignCompleted:                 false,
		ourProveBodyCompleted:            false,
		downstreamSubmitProveBodyCheckOK: false,
		msgFeedObjs:                      make([]event.Subscription, 0),
		logFeedObjs:                      make([]event.Subscription, 0),
		logFeeds:                         event.Feed{},
		clearTimeout:                     make(chan bool, 1), // 清除超时监控
		successedBackCall:                make([]func(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill), 0),
	}
	// Automatic timeout processing
	ins.startTimeoutControl()
	return ins
}

// Timeout monitoring
func (c *ChannelPayActionInstance) startTimeoutControl() {
	go func() {
		defer close(c.clearTimeout)
		select {
		case <-c.clearTimeout:
			return // end
		case <-time.NewTicker(time.Second * 20).C:
			// All payments must be completed within 30 seconds
			c.statusUpdateMux.Lock()
			if c.isBeDestroyed {
				c.statusUpdateMux.Unlock()
				return // Destroyed
			}
			c.statusUpdateMux.Unlock()
			// Timeout notification
			c.logError("Waiting over 20 seconds, payment has canceled by timeout.")
			c.Destroy() // Destruction
		}
	}()
}

// Complete all, destroy all resources
func (c *ChannelPayActionInstance) Destroy() {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	c.destroyUnsafe()
}

// Callback of successful payment
func (c *ChannelPayActionInstance) SetSuccessedBackCall(fc func(newbill *channel.OffChainCrossNodeSimplePaymentReconciliationBill)) {
	c.successedBackCall = append(c.successedBackCall, fc)
}

// Set local service node
func (c *ChannelPayActionInstance) SetLocalServicerNode(localnode *payroutes.PayRelayNode) {
	c.localServicerNode = localnode
}

func (c *ChannelPayActionInstance) destroyUnsafe() {

	if c.isBeDestroyed {
		return // Has been destroyed
	}
	c.log("payment instance destroyed.")
	c.isBeDestroyed = true // Destroy mark
	c.clearTimeout <- true // Clear timeout monitoring
	// Terminate message subscription
	for _, v := range c.msgFeedObjs {
		v.Unsubscribe()
	}
	// Terminate log subscription
	c.logFeeds.Send(&PayActionLog{
		IsEnd: true, // 日志结束
	})
	for _, v := range c.logFeedObjs {
		v.Unsubscribe()
	}
	// Relay service provider node disconnected
	if c.upstreamSide != nil {
		if c.downstreamSide != nil || c.collectCustomer != nil {
			c.upstreamSide.ChannelSide.WsConn.Close()
		}
	}
	if c.downstreamSide != nil {
		if c.upstreamSide != nil || c.payCustomer != nil {
			c.downstreamSide.ChannelSide.WsConn.Close()
		}
	}
	// Remove state exclusivity
	if c.upstreamSide != nil {
		c.upstreamSide.ClearBusinessExclusive() // Exclusive release
	}
	if c.downstreamSide != nil {
		//c.downstreamSide.ChannelSide.WsConn.Close()
		c.downstreamSide.ClearBusinessExclusive() // Exclusive release
	}
	if c.payCustomer != nil {
		c.payCustomer.ClearBusinessExclusive() // Exclusive release
	}
	if c.collectCustomer != nil {
		c.collectCustomer.ClearBusinessExclusive() // Exclusive release
	}
}

// Subscription log
func (c *ChannelPayActionInstance) SubscribeLogs(logschan chan *PayActionLog) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	subobj := c.logFeeds.Subscribe(logschan)
	c.logFeedObjs = append(c.logFeedObjs, subobj)
	//c.log("log subscribed")
}

// journal
func (c *ChannelPayActionInstance) log(con string) {
	c.logFeeds.Send(&PayActionLog{
		Content: con,
	})
}
func (c *ChannelPayActionInstance) logSuccess(con string) {
	c.logFeeds.Send(&PayActionLog{
		IsSuccess: true,
		Content:   "[SUCCESS] " + con,
	})
}
func (c *ChannelPayActionInstance) logError(con string) {
	c.logFeeds.Send(&PayActionLog{
		IsError: true,
		Content: "[ERROR] " + con,
	})
}

// Judge my node type 1,2,3
func (c *ChannelPayActionInstance) GetPayActionInstanceType() PayActionInstanceType {
	havup := c.payCustomer != nil || c.upstreamSide != nil
	havdown := c.collectCustomer != nil || c.downstreamSide != nil
	if havup && havdown {
		return RelayNode // Both upstream and downstream are intermediate nodes
	}
	if havup {
		return CollectSide // Payee
	} else {
		return PaySide // Payment end
	}

}

// Set connection parties
func (c *ChannelPayActionInstance) SetUpstreamSide(side *RelayPaySettleNoder) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	c.upstreamSide = side
}
func (c *ChannelPayActionInstance) SetDownstreamSide(side *RelayPaySettleNoder) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	c.downstreamSide = side
}
func (c *ChannelPayActionInstance) SetPayCustomer(user *Customer) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	c.payCustomer = user
}
func (c *ChannelPayActionInstance) SetCollectCustomer(user *Customer) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	c.collectCustomer = user
}

// Check address
func (c *ChannelPayActionInstance) checkOurAddress(addr fields.Address) bool {
	if c.ourMustSignAddresses != nil {
		for _, v := range c.ourMustSignAddresses {
			if v.Equal(addr) {
				return true
			}
		}
	}
	return false
}

// Set signing machine
func (c *ChannelPayActionInstance) SetSignatureMachine(machine DataSourceOfSignatureMachine) {
	c.signMachine = machine
}

// Set signature address
func (c *ChannelPayActionInstance) SetMustSignAddresses(addrs []fields.Address) {
	c.ourMustSignAddresses = addrs
}

// Check whether you can sign. If you can, complete the signature directly
func (c *ChannelPayActionInstance) checkMaybeCanDoSign() (bool, error) {

	// Have I signed all
	if c.ourSignCompleted {
		return true, nil // Sign only once
	}

	// Check the completion of the statement
	var mustSignList = make([]fields.Address, 0)
	var prevlast fields.Address = nil
	for i, v := range c.billDocuments.ProveBodys.ProveBodys {
		chr := c.billDocuments.ChainPayment.ChannelTransferProveHashHalfCheckers[i]
		if v == nil || chr == nil {
			//fmt.Println("checkMaybeCanDoSign: ProveBodys not complete.", i)
			return false, nil
			// The transaction body is not ready
		}
		// inspect
		// If you pay from right to left
		var a1, a2 = v.LeftAddress, v.RightAddress
		if uint8(v.PayDirection) == channel.ChannelTransferDirectionHacashRightToLeft {
			a2, a1 = a1, a2
		}
		if prevlast.Equal(a1) {
			mustSignList = append(mustSignList, a2)
		} else {
			mustSignList = append(mustSignList, a1, a2)
		}
		prevlast = a2 // record
	}

	// Check whether the statement detection passes
	if c.downstreamSubmitProveBodyCheckOK == false {
		return false, fmt.Errorf("Downstream submit prove body NOT check OK.")
	}

	// Check signature
	for i := len(mustSignList) - 1; i >= 0; i-- {
		one := mustSignList[i]
		if c.checkOurAddress(one) {
			break // Inspection completed
		}
		// Check whether the downstream signature is completed
		cke := c.billDocuments.ChainPayment.CheckOneAddressSign(one)
		if cke != nil {
			// Downstream signature check failed
			return false, nil // No error returned, wait for the next check
		}
	}

	// Execute my signature
	mysigns, e := c.signMachine.CheckPaydocumentAndFillNeedSignature(c.billDocuments, c.ourMustSignAddresses)
	if e != nil {
		// Return signature error
		nodemark := ""
		if c.localServicerNode != nil {
			nodemark = "(relay node: " + c.localServicerNode.IdentificationName.Value() + ") "
		}
		return false, fmt.Errorf("%ssignature fail: %s", nodemark, e.Error())
	}

	// Broadcast my signature
	sgmsg := &protocol.MsgBroadcastChannelStatementSignature{
		TransactionDistinguishId: c.transactionDistinguishId,
		Signs:                    *mysigns,
	}
	// Broadcast message
	c.BroadcastMessage(sgmsg)
	// Fill in my signature
	for _, v := range mysigns.Signs {
		e := c.billDocuments.ChainPayment.FillSignByPosition(v)
		if e != nil {
			// Failed to return fill signature
			return false, fmt.Errorf("fill signature fail: %s", e.Error())
		}
	}
	c.log("my signnature already broadcast")
	// I signed successfully
	c.ourSignCompleted = true // sign
	return true, nil
}

// Create statement
func (c *ChannelPayActionInstance) createMyProveBodyByRemotePay(collectAmt *fields.Amount) (*channel.ChannelChainTransferProveBodyInfo, error) {

	// Find the channel
	var chanSide *ChannelSideConn = nil
	if c.upstreamSide != nil {
		chanSide = c.upstreamSide.ChannelSide
	} else if c.payCustomer != nil {
		chanSide = c.payCustomer.ChannelSide
	} else {
		return nil, fmt.Errorf("not find up side ChannelSide ptr.")
	}

	// Check capacity
	amtcap := chanSide.GetChannelCapacityAmountOfRemote()
	if amtcap.LessThan(collectAmt) {
		return nil, fmt.Errorf("up side channel capacity balance not enough.")
	}

	// Create statement
	chaninfo := chanSide.ChannelInfo
	reuseversion := uint32(chaninfo.ReuseVersion)
	billautonumber := uint64(1)
	lastbill := chanSide.GetReconciliationBill()
	//fmt.Printf("1:%p, 2:%p\n", chanSide, lastbill)
	oldleftamt := chaninfo.LeftAmount
	oldrightamt := chaninfo.RightAmount
	if lastbill != nil {
		//fmt.Println("chanSide.GetReconciliationBill()", lastbill.GetAutoNumber())
		oldleftamt = lastbill.GetLeftBalance()
		oldrightamt = lastbill.GetRightBalance()
		blrn, blan := lastbill.GetReuseVersionAndAutoNumber()
		if reuseversion != blrn {
			return nil, fmt.Errorf("Channel Reuse Version %d in last bill and %d in channel info not match.", blrn, reuseversion)
		}
		billautonumber = uint64(blan) + 1 // Autoincrement
	}
	// direction
	paydirection := channel.ChannelTransferDirectionHacashRightToLeft
	oldpayamt := oldrightamt
	oldcollectamt := oldleftamt
	remoteisleft := chanSide.RemoteAddressIsLeft()
	if remoteisleft {
		paydirection = channel.ChannelTransferDirectionHacashLeftToRight // Always paid to me
		oldpayamt, oldcollectamt = oldcollectamt, oldpayamt              // reverse
	}
	// Calculate latest allocation
	newpayamt, e := oldpayamt.Sub(collectAmt) // 支付端扣除
	if e != nil {
		return nil, fmt.Errorf("pay error: %s", e.Error())
	}
	if newpayamt == nil || newpayamt.IsNegative() {
		return nil, fmt.Errorf("pay error: balance not enough.")
	}
	newcollectamt, e := oldcollectamt.Add(collectAmt) // 收款段增加
	if e != nil {
		return nil, fmt.Errorf("collect error: %s", e.Error())
	}
	if newcollectamt == nil || newpayamt.IsNegative() {
		return nil, fmt.Errorf("collect error: balance is Negative.")
	}
	newleftamt := newcollectamt
	newrightamt := newpayamt
	if remoteisleft {
		newleftamt, newrightamt = newrightamt, newleftamt // reverse
	}
	//fmt.Println("createMyProveBodyByRemotePay: billautonumber=", billautonumber)
	// establish
	body := channel.CreateEmptyProveBody(chanSide.ChannelId)
	body.ReuseVersion = fields.VarUint4(reuseversion)
	body.BillAutoNumber = fields.VarUint8(billautonumber)
	body.PayDirection = fields.VarUint1(paydirection)
	body.PayAmount = *collectAmt
	body.LeftBalance = *newleftamt
	body.RightBalance = *newrightamt
	body.LeftAddress = chaninfo.LeftAddress
	body.RightAddress = chaninfo.RightAddress

	// return
	return body, nil
}

// Do you want to report my trading body
func (c *ChannelPayActionInstance) checkMaybeReportMyProveBody(msg *protocol.MsgRequestInitiatePayment) (bool, error) {
	if c.ourProveBodyCompleted {
		return true, nil // Prevent duplicate reporting
	}

	var e error
	var bodyindex int = 0
	var bodyinfo *channel.ChannelChainTransferProveBodyInfo = nil

	// First report as final payee
	if msg != nil && (c.downstreamSide == nil && c.collectCustomer == nil) {
		upside := c.upstreamSide
		if upside == nil {
			return false, fmt.Errorf("upstreamSide is nil")
		}
		// Create statement
		bodyindex = c.channelLength - 1
		bodyinfo, e = c.createMyProveBodyByRemotePay(&msg.PayAmount)
		if e != nil {
			return false, fmt.Errorf("create my prove body by remote pay error: %s", e.Error())
		}
		// I am the final payee, and I do not need to check the statement submitted by others
		c.downstreamSubmitProveBodyCheckOK = true // ok
	} else if c.upstreamSide == nil && c.payCustomer == nil {
		// I am the source payer and do not need to generate a transaction body. It is generated by my service provider
		c.ourProveBodyIndex = 0
		c.ourProveBodyCompleted = true
		return true, nil

	} else {

		// Check whether the downstream has filled the report
		var downSide *ChannelSideConn = nil
		if c.downstreamSide != nil {
			downSide = c.downstreamSide.ChannelSide
		} else if c.collectCustomer != nil {
			downSide = c.collectCustomer.ChannelSide
		} else {
			return false, fmt.Errorf("channel pay down stream side not find")
		}
		// Downstream payment
		var downSideColletAmt *fields.Amount = nil
		var payaddr *fields.Address = nil
		bdlist := c.billDocuments.ProveBodys.ProveBodys
		for i := len(bdlist) - 1; i >= 0; i-- {
			one := bdlist[i]
			if one != nil && one.ChannelId.Equal(downSide.ChannelId) {
				bodyindex = i - 1
				downSideColletAmt = &one.PayAmount
				payaddr = &one.LeftAddress
				if uint8(one.PayDirection) == channel.ChannelTransferDirectionHacashRightToLeft {
					payaddr = &one.RightAddress
				}
			}
		}
		if downSideColletAmt == nil {
			// The downstream has not broadcast the statement, waiting for the next test
			// No error returned
			return false, nil
		}
		// Check if the payment segment is me
		if payaddr.NotEqual(downSide.OurAddress) {
			return false, fmt.Errorf("pay address %s is not our address %s.",
				payaddr.ToReadable(), downSide.OurAddress.ToReadable())
		}
		// Check whether I am a payer
		var upSide *ChannelSideConn = nil
		if c.upstreamSide != nil {
			upSide = c.upstreamSide.ChannelSide
		} else if c.payCustomer != nil {
			upSide = c.payCustomer.ChannelSide
		} else {
			// 没有上游，我就是付款端，不用广播什么，等待签名即可
			return false, nil
		}
		if bodyindex < 0 {
			// 没有上游，我就是付款端，不用广播什么，等待签名即可
			return false, nil
		}
		// Plus handling charges
		// Plus the service charge of my own node
		lcsnode := c.localServicerNode
		if lcsnode == nil {
			return false, fmt.Errorf("local servicer node is nil.")
		}
		// Service Charge
		appendfee := lcsnode.PredictFeeForPay(downSideColletAmt)
		newpayamt, e := downSideColletAmt.Add(appendfee)
		if e != nil {
			return false, fmt.Errorf("add predict fee for pay error: %s", e.Error())
		}
		// Create statement
		bodyinfo, e = c.createMyProveBodyByRemotePay(newpayamt)
		if e != nil {
			return false, fmt.Errorf("create my prove body of channel id %s by remote pay error: %s", upSide.ChannelId.ToHex(), e.Error())
		}
	}

	if bodyinfo == nil {
		return false, nil
	}

	// My aisle location
	c.ourProveBodyIndex = bodyindex

	// journal
	c.log(fmt.Sprintf("prove body %d/%d created and broadcast...", bodyindex+1, c.channelLength))
	// Fill my ticket in the specified location
	// Fill ticket
	c.fillBillDocumentsUseProveBody(bodyindex, bodyinfo)

	msgbd := &protocol.MsgBroadcastChannelStatementProveBody{
		TransactionDistinguishId: c.transactionDistinguishId,
		ProveBodyIndex:           fields.VarUint1(bodyindex),
		ProveBodyInfo:            bodyinfo,
	}
	c.BroadcastMessage(msgbd)      // Broadcast message
	c.ourProveBodyCompleted = true // Report successful
	return true, nil

}

// Initializing the creation of transaction tickets through the message
func (c *ChannelPayActionInstance) InitCreateEmptyBillDocumentsByInitPayMsg(msg *protocol.MsgRequestInitiatePayment) error {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()
	c.log(fmt.Sprintf("init: create pay bill documents for transfer %s to %s",
		msg.PayAmount.ToFinString(), msg.PayeeChannelAddr.Value()))
	// Channel length
	var paychanlen = int(msg.TargetPath.NodeIdCount) + 1
	c.channelLength = paychanlen
	c.allSignsCompleted = false
	c.transactionDistinguishId = msg.TransactionDistinguishId
	// Transaction notes
	bodys := make([]*channel.ChannelChainTransferProveBodyInfo, paychanlen)
	for i := 0; i < paychanlen; i++ {
		bodys[i] = nil
	}
	c.billDocuments = &channel.ChannelPayCompleteDocuments{
		ProveBodys: &channel.ChannelPayProveBodyList{
			Count:      fields.VarUint1(paychanlen),
			ProveBodys: bodys,
		},
		ChainPayment: &channel.OffChainFormPaymentChannelTransfer{
			Timestamp:                            msg.Timestamp,
			OrderNoteHashHalfChecker:             msg.OrderNoteHashHalfChecker,
			MustSignCount:                        0,
			MustSignAddresses:                    make([]fields.Address, 0),
			ChannelCount:                         fields.VarUint1(paychanlen),
			ChannelTransferProveHashHalfCheckers: make([]fields.HashHalfChecker, paychanlen),
			MustSigns:                            make([]fields.Sign, 0),
		},
	}

	c.log("bill documents created...")

	// Message cache
	c.payInitMsg = msg

	// If I am the final payee, I need to be the first to report my transaction
	_, e := c.checkMaybeReportMyProveBody(msg)
	if e == nil {
	}
	return e
}

// Get the other side connection for forwarding messages
func (c *ChannelPayActionInstance) getUpOrDownStreamNegativeDirection(upOrDownStream bool) *websocket.Conn {
	var conn *websocket.Conn = nil
	if upOrDownStream {
		if c.downstreamSide != nil {
			conn = c.downstreamSide.ChannelSide.WsConn
		}
		if c.collectCustomer != nil {
			conn = c.collectCustomer.ChannelSide.WsConn
		}
	} else {
		if c.upstreamSide != nil {
			conn = c.upstreamSide.ChannelSide.WsConn
		}
		if c.payCustomer != nil {
			conn = c.payCustomer.ChannelSide.WsConn
		}
	}
	return conn
}

// Broadcast message
func (c *ChannelPayActionInstance) BroadcastMessage(msg protocol.Message) {
	wsconn1 := c.getUpOrDownStreamNegativeDirection(true)
	wsconn2 := c.getUpOrDownStreamNegativeDirection(false)
	if wsconn1 != nil {
		protocol.SendMsg(wsconn1, msg)
	}
	if wsconn2 != nil {
		protocol.SendMsg(wsconn2, msg)
	}
}

// Start a side message subscription
func (c *ChannelPayActionInstance) StartOneSideMessageSubscription(upOrDownStream bool, side *ChannelSideConn) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	c.log("start subscribe message, wait next...")
	msgchanobj := make(chan protocol.Message, 2)
	subobj := side.SubscribeMessage(msgchanobj)
	c.msgFeedObjs = append(c.msgFeedObjs, subobj)
	// Start listening
	go func() {
		for {
			select {
			case msgobj := <-msgchanobj:
				if msgobj == nil {
					return // Stop listening
				}
				//fmt.Println("msgobj := <-msgchanobj:", msgobj.Type())
				var e error = nil
				switch msgobj.Type() {
				// Wrong arrival
				case protocol.MsgTypeBroadcastChannelStatementError:
					c.channelPayErrorArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementError))
				// Arrival of statement
				case protocol.MsgTypeBroadcastChannelStatementProveBody:
					e = c.channelPayProveBodyArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementProveBody))
				// Signature arrival
				case protocol.MsgTypeBroadcastChannelStatementSignature:
					e = c.channelPaySignatureArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementSignature))
				// Payment succeeded
				case protocol.MsgTypeBroadcastChannelStatementSuccessed:
					c.channelPaySuccessArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementSuccessed))
				}
				// error handling
				if e != nil {
					// Broadcast error
					errkind := ""
					if c.localServicerNode != nil {
						errkind = c.localServicerNode.IdentificationName.Value()
					} else if c.downstreamSide == nil && c.collectCustomer == nil {
						errkind = "collect side"
					} else if c.upstreamSide == nil && c.payCustomer == nil {
						errkind = "pay side"
					}
					errorMsg := "(" + errkind + ") " + "do payment got error: " + e.Error()
					c.logError(errorMsg)
					// Broadcast error
					msg := &protocol.MsgBroadcastChannelStatementError{
						ErrCode: 1,
						ErrTip:  fields.CreateStringMax65535(errorMsg),
					}
					c.BroadcastMessage(msg)
					// End all
					c.Destroy()
				}
			case <-subobj.Err():
				return // Stop listening
			}
		}
	}()

}

// Start message listening on one side
func (c *ChannelPayActionInstance) delete___startOneSideMessageListen(upOrDownStream bool, conn *websocket.Conn) {
	go func() {
		for {
			// Read message
			msgobj, _, e := protocol.ReceiveMsg(conn)
			if e != nil {
				break // error
			}
			if msgobj != nil {
				var e error = nil
				switch msgobj.Type() {
				// Wrong arrival
				case protocol.MsgTypeBroadcastChannelStatementError:
					c.channelPayErrorArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementError))
				// Arrival of statement
				case protocol.MsgTypeBroadcastChannelStatementProveBody:
					e = c.channelPayProveBodyArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementProveBody))
				// Signature arrival
				case protocol.MsgTypeBroadcastChannelStatementSignature:
					e = c.channelPaySignatureArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementSignature))
				// Payment succeeded
				case protocol.MsgTypeBroadcastChannelStatementSuccessed:
					c.channelPaySuccessArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementSuccessed))
				}
				// error handling
				if e != nil {
					// Broadcast error
					errkind := ""
					if c.localServicerNode != nil {
						errkind = c.localServicerNode.IdentificationName.Value()
					} else if c.downstreamSide == nil && c.collectCustomer == nil {
						errkind = "collect side"
					} else if c.upstreamSide == nil && c.payCustomer == nil {
						errkind = "pay side"
					}
					msg := &protocol.MsgBroadcastChannelStatementError{
						ErrCode: 1,
						ErrTip:  fields.CreateStringMax65535("<" + errkind + "> " + e.Error()),
					}
					c.BroadcastMessage(msg)
				}
			}
		}
		// An error occurred disconnecting

	}()
}

// Fill in the transaction bill, and the statement arrives
func (c *ChannelPayActionInstance) channelPayProveBodyArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementProveBody) error {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// fill
	checkerlist := c.billDocuments.ChainPayment.ChannelTransferProveHashHalfCheckers
	cid := int(msg.ProveBodyIndex)
	if len(checkerlist) < cid {
		return fmt.Errorf("ProveBodyIndex overflow pay channel chain length.")
	}

	// Check whether the serial number of the transaction entity reported by the opposite party is consistent with my bill
	var downside *ChannelSideConn = nil
	if c.downstreamSide != nil {
		downside = c.downstreamSide.ChannelSide
	} else if c.collectCustomer != nil {
		downside = c.collectCustomer.ChannelSide
	}
	if downside != nil && downside.ChannelId.Equal(msg.ProveBodyInfo.ChannelId) {
		// Check whether the reported transaction reconciliation body meets the requirements
		ckn1 := uint32(downside.ChannelInfo.ReuseVersion)
		ckn2 := uint64(1)
		oldamtl := downside.ChannelInfo.LeftAmount
		oldamtr := downside.ChannelInfo.RightAmount
		oldbill := downside.GetReconciliationBill()
		//fmt.Printf("1:%p, 2:%p\n", downside, oldbill)
		if oldbill != nil {
			ckn1 = oldbill.GetReuseVersion()
			ckn2 = oldbill.GetAutoNumber() + 1
			oldamtl = oldbill.GetLeftBalance()
			oldamtr = oldbill.GetRightBalance()
		}
		num1 := msg.ProveBodyInfo.GetReuseVersion()
		num2 := msg.ProveBodyInfo.GetAutoNumber()
		if ckn1 != num1 {
			return fmt.Errorf("ProveBody arrive check reuseVersion need %d but got %d.", ckn1, num1)
		}
		if ckn2 != num2 {
			return fmt.Errorf("ProveBody arrive check autoNumber need %d but got %d.", ckn2, num2)
		}
		// Total number of inspection channels
		needcap := downside.ChannelInfo.GetLeftAndRightTotalAmount()
		namtl := msg.ProveBodyInfo.GetLeftBalance()
		namtr := msg.ProveBodyInfo.GetRightBalance()
		newcap, _ := namtl.Add(&namtr)
		if newcap.NotEqual(needcap) {
			oldamtl.ToFinString()
			oldamtr.ToFinString()
			return fmt.Errorf("ProveBody arrive check LeftAndRightTotalAmount need %d but got %d.",
				needcap.ToFinString(), newcap.ToFinString())
		}
		// Mark check complete
		c.downstreamSubmitProveBodyCheckOK = true
	}

	// As the source payer, I check the content and payment amount of the transaction body
	if c.upstreamSide == nil && c.payCustomer == nil && msg.ProveBodyIndex == 0 {
		// Check amount
		if c.payInitMsg == nil {
			return fmt.Errorf("c.payInitMsg is nil")
		}
		feeup := c.payInitMsg.HighestAcceptanceFee
		feemax, _ := feeup.Add(&feeup)
		amtmaxset, _ := c.payInitMsg.PayAmount.Add(feemax)
		amtbd := msg.ProveBodyInfo.PayAmount
		amtinit := c.payInitMsg.PayAmount
		if amtbd.LessThan(&amtinit) {
			return fmt.Errorf("payment prove body amount %s less than init msg amount %s",
				amtbd.ToFinString(), amtinit.ToFinString())
		}
		if amtbd.MoreThan(amtmaxset) {
			return fmt.Errorf("payment prove body amount %s more than max set amount %s",
				amtbd.ToFinString(), amtmaxset.ToFinString())
		}
		// Payment amount check succeeded
	}

	// journal
	c.log(fmt.Sprintf("prove body %d/%d arrived...", cid+1, len(checkerlist)))

	// Fill ticket
	c.fillBillDocumentsUseProveBody(cid, msg.ProveBodyInfo)

	// Forward the transaction body to the other party
	otherside := c.getUpOrDownStreamNegativeDirection(upOrDownStream)
	if otherside != nil { //
		// forward
		protocol.SendMsg(otherside, msg)
	}

	// Do you want to report my trading body
	_, e := c.checkMaybeReportMyProveBody(nil)
	if e != nil {
		return e
	}

	// Can I sign
	//fmt.Println("channelPayProveBodyArrive   _, e = c.checkMaybeCanDoSign()")
	_, e = c.checkMaybeCanDoSign()
	if e != nil {
		return e
	}

	// ok
	return nil

}

// Signature data arrival
func (c *ChannelPayActionInstance) channelPaySignatureArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementSignature) error {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// Fill in the signature, do not check, ignore the error
	addrs := make([]string, msg.Signs.Count)
	for i, v := range msg.Signs.Signs {
		addrs[i] = v.GetAddress().ToReadable()
		e := c.billDocuments.ChainPayment.FillSignByPosition(v)
		if e != nil {
			return e
		}
	}

	// Forward signature to another party
	otherside := c.getUpOrDownStreamNegativeDirection(upOrDownStream)
	if otherside != nil {
		// forward
		protocol.SendMsg(otherside, msg)
	}

	// journal
	c.log(fmt.Sprintf("received signatures of %s", strings.Join(addrs, ", ")))

	// Can I sign it myself
	_, e := c.checkMaybeCanDoSign()
	if e != nil {
		return e
	}

	// If I am the last payee, when the last signature arrives, I am responsible for broadcasting the message of successful completion of payment
	if c.downstreamSide == nil && c.collectCustomer == nil {
		e := c.billDocuments.ChainPayment.CheckMustAddressAndSigns()
		if e == nil {
			// Success log
			c.logSuccess(fmt.Sprintf("collect successfully finished at %s.", time.Now().Format("15:04:05.999")))
			// All signatures are completed and broadcast completion message
			okmsg := &protocol.MsgBroadcastChannelStatementSuccessed{
				SuccessTip: fields.CreateStringMax65535(""),
			}
			c.BroadcastMessage(okmsg)
			// Destroy payment operation package
			c.destroyUnsafe()
			// Callback of payment completion
			go c.doCallbackOfSuccessed() // Callback
		}
	}

	// ok
	return nil
}

// Wrong arrival
func (c *ChannelPayActionInstance) channelPayErrorArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementError) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	fmt.Println("channelPayErrorArrive:", msg.ErrTip.Value())

	// Forward errors to another party
	otherside := c.getUpOrDownStreamNegativeDirection(upOrDownStream)
	if otherside != nil {
		protocol.SendMsg(otherside, msg)
	}

	// Print error log
	c.logError(msg.ErrTip.Value())

	// End all
	c.destroyUnsafe()
}

// Payment completed successfully
func (c *ChannelPayActionInstance) channelPaySuccessArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementSuccessed) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// Check signature
	e := c.billDocuments.ChainPayment.CheckMustAddressAndSigns()
	if e != nil {
		// The log check failed but the payment was successfully completed
		c.logError(fmt.Sprintf("sign check failed but got successed message."))
		c.destroyUnsafe()
		return
	}

	// Successfully forwarded to upstream, one-way message
	otherside := c.getUpOrDownStreamNegativeDirection(false)
	if otherside != nil {
		protocol.SendMsg(otherside, msg)
	}

	// journal
	c.logSuccess(fmt.Sprintf("payment finished successfully at %s.", time.Now().Format("15:04:05.999")))

	// End all
	c.destroyUnsafe()

	// Callback of payment completion
	go c.doCallbackOfSuccessed() // Callback

}

// Payment success callback
func (c *ChannelPayActionInstance) doCallbackOfSuccessed() {

	doCall := func(provebody *channel.ChannelChainTransferProveBodyInfo) {
		bill := &channel.OffChainCrossNodeSimplePaymentReconciliationBill{
			ChannelChainTransferTargetProveBody: *provebody,
			ChannelChainTransferData:            *c.billDocuments.ChainPayment,
		}
		for _, f := range c.successedBackCall {
			f(bill) // Callback
		}
	}

	// Find reconciliation notes
	bdlist := c.billDocuments.ProveBodys.ProveBodys
	// call
	provebody := bdlist[c.ourProveBodyIndex]
	go doCall(provebody)

	// If I am an intermediate node, I will call the successful statement again
	if RelayNode == c.GetPayActionInstanceType() {
		// Recall the next channel position again
		provebody = bdlist[c.ourProveBodyIndex+1]
		go doCall(provebody)
	}
}

// Fill ticket
func (c *ChannelPayActionInstance) fillBillDocumentsUseProveBody(bodyIndex int, proveBodyInfo *channel.ChannelChainTransferProveBodyInfo) {
	// fill
	c.billDocuments.ProveBodys.ProveBodys[bodyIndex] = proveBodyInfo // fill
	c.billDocuments.ChainPayment.ChannelTransferProveHashHalfCheckers[bodyIndex] = proveBodyInfo.GetSignStuffHashHalfChecker()
	// Add address
	addrlist := make([]fields.Address, 2)
	addrlist[0] = proveBodyInfo.GetLeftAddress()
	addrlist[1] = proveBodyInfo.GetRightAddress()
	addrlen, addrs := fields.CleanAddressListByCharacterSort(c.billDocuments.ChainPayment.MustSignAddresses, addrlist)
	// Fill ticket
	c.billDocuments.ChainPayment.MustSignCount = fields.VarUint1(addrlen)
	c.billDocuments.ChainPayment.MustSignAddresses = addrs
	allsigns := make([]fields.Sign, addrlen)
	for i := 0; i < int(addrlen); i++ {
		allsigns[i] = fields.CreateEmptySign()
	}
	c.billDocuments.ChainPayment.MustSigns = allsigns

	// Successfully completed
	return
}
