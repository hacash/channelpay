package client

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"github.com/hacash/mint/event"
	"github.com/hacash/node/websocket"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

const (
	billDir string = "hacash_channel_pay_bill_data"
)

/**
 * 通道链支付用户端
 */
type ChannelPayUser struct {
	changeMux sync.Mutex

	// Account address
	selfAcc  *account.Account
	selfAddr *protocol.ChannelAccountAddress
	chanInfo *protocol.RpcDataChannelInfo

	// Local reconciliation bill
	localLatestReconciliationBalanceBill channel.ReconciliationBalanceBill

	// Waiting for statement
	waitRealtimeReconciliation *channel.OffChainFormPaymentChannelRealtimeReconciliation

	// Channel chain
	servicerStreamSide *chanpay.RelayPaySettleNoder

	// Message subscription
	msgSubObj event.Subscription

	// Has it been closed
	isClosed bool
}

func CreateChannelPayUser(acc *account.Account, addr *protocol.ChannelAccountAddress, cinfo *protocol.RpcDataChannelInfo) *ChannelPayUser {
	return &ChannelPayUser{
		selfAcc:  acc,
		selfAddr: addr,
		chanInfo: cinfo,
		isClosed: false,
	}
}

// Check whether the collection channel is occupied
func (c *ChannelPayUser) IsInBusinessExclusive() bool {
	return c.servicerStreamSide.IsInBusinessExclusive()
}

// Where state exclusive
func (c *ChannelPayUser) StartBusinessExclusive() bool {
	return c.servicerStreamSide.StartBusinessExclusive()
}

// Remove state exclusivity
func (c *ChannelPayUser) ClearBusinessExclusive() {
	c.servicerStreamSide.ClearBusinessExclusive()
}

// Start sending heartbeat packets
func (c *ChannelPayUser) StartHeartbeat() {
	for {
		// A heartbeat packet in 15 seconds
		time.Sleep(time.Second * 14)
		c.changeMux.Lock()
		if c.isClosed {
			c.changeMux.Unlock()
			return // end
		}
		c.changeMux.Unlock()
		// Send heartbeat packet, ignore error
		//fmt.Println("protocol.SendMsg(c.servicerStreamSide.ChannelSide.WsConn, &protocol.MsgHeartbeat{})")
		protocol.SendMsg(c.servicerStreamSide.ChannelSide.WsConn, &protocol.MsgHeartbeat{})
		// Determine the heartbeat returned by the server
		lastBeatTime := c.servicerStreamSide.ChannelSide.GetLastestHeartbeatTime()
		tnck := time.Now().Unix() - 60
		if lastBeatTime.Unix() < tnck {
			// No heartbeat packet from the server has been received for 60 seconds, and it will be disconnected automatically
			c.Logout() // sign out
		}
	}
}

// sign out
func (c *ChannelPayUser) IsClosed() bool {
	return c.isClosed // Have you exited or closed
}

func (c *ChannelPayUser) Logout() {
	c.changeMux.Lock()
	defer c.changeMux.Unlock()
	if c.isClosed {
		return // Has exited
	}
	c.isClosed = true
	if c.servicerStreamSide != nil {
		wsconn := c.servicerStreamSide.ChannelSide.WsConn
		// Send exit message
		protocol.SendMsg(wsconn, &protocol.MsgCustomerLogout{
			PostBack: fields.CreateStringMax255(""),
		})
		wsconn.Close()
	}
	// Close subscription
	if c.msgSubObj != nil {
		c.msgSubObj.Unsubscribe()
		c.msgSubObj = nil
	}
}

func billfilepath(channelID fields.ChannelId) string {
	datadir := sys.AbsDir(billDir)
	os.MkdirAll(datadir, 0777)
	fname := path.Join(datadir, fmt.Sprintf("bill_%s.dat", channelID.ToHex()))
	return fname
}

func LoadBillFromDisk(channelID fields.ChannelId) (channel.ReconciliationBalanceBill, error) {
	var bill channel.ReconciliationBalanceBill = nil
	// Read local directory
	fname := billfilepath(channelID)
	fbts, e := ioutil.ReadFile(fname)
	if e == nil || len(fbts) > 0 {
		// Parse file
		bill, _, e = channel.ParseReconciliationBalanceBillByPrefixTypeCode(fbts, 0)
	}
	// return
	return bill, nil
}

// Delete local ticket
func (c *ChannelPayUser) DeleteLastBillOnDisk() error {
	// Read local directory
	fname := billfilepath(c.selfAddr.ChannelId)
	c.localLatestReconciliationBalanceBill = nil
	// Delete file
	return os.Remove(fname)
}

// Read local reconciliation bills
func (c *ChannelPayUser) LoadLastBillFromDisk() (channel.ReconciliationBalanceBill, error) {
	var bill channel.ReconciliationBalanceBill = nil
	// Read local directory
	ldbill, e := LoadBillFromDisk(c.selfAddr.ChannelId)
	if e == nil {
		bill = ldbill
	}
	c.localLatestReconciliationBalanceBill = bill
	// return
	return bill, nil
}

// Save reconciliation bill
func (c *ChannelPayUser) SaveLastBillToDisk(bill channel.ReconciliationBalanceBill) error {
	if bill == nil {
		return fmt.Errorf("bill is nil")
	}
	// Local directory
	fname := billfilepath(c.selfAddr.ChannelId)
	fbts, e := channel.SerializeReconciliationBalanceBillWithPrefixTypeCode(bill)
	if e == nil || len(fbts) > 0 {
		// preservation
		e = ioutil.WriteFile(fname, fbts, 0777)
		if e != nil {
			return e // Return error
		}
	}
	c.localLatestReconciliationBalanceBill = bill
	// ok
	return nil
}

// Obtain reconciliation bills remotely after login
func (c *ChannelPayUser) GetReconciliationBalanceBillAfterLoginFromRemote() channel.ReconciliationBalanceBill {
	if c.servicerStreamSide != nil {
		return c.servicerStreamSide.ChannelSide.LatestReconciliationBalanceBill
	}
	return nil // non-existent
}

// Connect to the server
func (c *ChannelPayUser) ConnectServicer(wsurl string) error {
	// dial
	wsconn, e := websocket.Dial(wsurl, "", "http://127.0.0.1")
	if e != nil {
		return fmt.Errorf("Connect servicer %s error: %s", wsurl, e.Error())
	}

	// Send identity
	resmsg, _, e := protocol.SendMsgForResponseTimeout(wsconn, &protocol.MsgLogin{
		ProtocolVersion: fields.VarUint2(protocol.LatestProtocolVersion),
		ChannelId:       c.selfAddr.ChannelId,
		CustomerAddress: c.selfAddr.Address,
		LanguageSet:     fields.CreateStringMax255("US"),
	}, 6)
	if e != nil {
		wsconn.Close()
		return fmt.Errorf("Send login msg error: %s", e.Error())
	}

	// Login error
	resmsgty := resmsg.Type()
	if resmsgty == protocol.MsgTypeError {
		wsconn.Close()
		msg := resmsg.(*protocol.MsgError)
		return fmt.Errorf("Do login error: %s", msg.ErrTip.Value())
	}

	// Message type
	if resmsgty != protocol.MsgTypeLoginCheckLastestBill {
		// Unsupported message type
		wsconn.Close()
		return fmt.Errorf("Unsupported message reply type = %d", resmsgty)
	}

	// Login successful
	billmsg := resmsg.(*protocol.MsgLoginCheckLastestBill)

	// Inspection protocol
	if uint16(billmsg.ProtocolVersion) > protocol.LatestProtocolVersion {
		wsconn.Close()
		return fmt.Errorf("You need to upgrade your client software, version %d => %d", protocol.LatestProtocolVersion, billmsg.ProtocolVersion)
	}
	if uint16(billmsg.ProtocolVersion) < protocol.LatestProtocolVersion {
		wsconn.Close()
		return fmt.Errorf("Your servicer does not support the protocol version %d of your client software", protocol.LatestProtocolVersion)
	}

	// Create upstreamchannelside
	csobj := chanpay.NewChannelSideByConn(wsconn)
	csobj.ChannelId = c.selfAddr.ChannelId
	csobj.ChannelInfo = c.chanInfo
	csobj.OurAddress = c.selfAddr.Address
	ourIsLeft := c.selfAddr.Address.Equal(c.chanInfo.LeftAddress)
	if ourIsLeft {
		csobj.RemoteAddress = c.chanInfo.RightAddress
	} else {
		csobj.RemoteAddress = c.chanInfo.LeftAddress
	}
	if billmsg.BillIsExistent.Check() {
		csobj.LatestReconciliationBalanceBill = billmsg.LastBill
	}

	// Channel end
	c.servicerStreamSide = chanpay.NewRelayPayNodeConnect(c.selfAddr.ServicerName.Value(), csobj.ChannelId, ourIsLeft, csobj)

	// Keep heartbeat alive
	go c.StartHeartbeat()

	// success
	return nil
}
