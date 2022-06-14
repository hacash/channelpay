package client

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
)

// Reconciliation return
func (c *ChannelPayClient) DealServicerRespondReconciliation(msg *protocol.MsgServicerRespondReconciliation) {

	c.user.changeMux.Lock()
	defer c.user.changeMux.Unlock()

	waitBill := c.user.waitRealtimeReconciliation
	if waitBill == nil {
		return // No list
	}
	tarAddr := account.NewAddressFromPublicKeyV0(msg.SelfSign.PublicKey)
	if waitBill.LeftAddress.Equal(tarAddr) {
		waitBill.LeftSign = msg.SelfSign
	} else {
		waitBill.RightSign = msg.SelfSign
	}
	// Check document signature
	if e := waitBill.CheckAddressAndSign(); e != nil {
		c.user.waitRealtimeReconciliation = nil // signature check failed
		return
	}

	// Update bill status
	c.updateReconciliationBalanceBill(waitBill)
}

// Start reconciliation
func (c *ChannelPayClient) InitiateReconciliation(bill *channel.OffChainCrossNodeSimplePaymentReconciliationBill) {

	//c.changeMux.Lock()
	//defer c.changeMux.Unlock()

	// state
	chanside := c.user.servicerStreamSide.ChannelSide
	chanside.StartBusinessExclusive() // State exclusive
	defer chanside.ClearBusinessExclusive()

	// Convert to statement
	recbill := bill.ConvertToRealtimeReconciliation()

	// Calculate our signature
	sign, _, e := recbill.FillTargetSignature(c.user.selfAcc)
	if e != nil {
		return // fail
	}

	// record
	c.user.waitRealtimeReconciliation = recbill

	// Send signature to the server
	conn := chanside.WsConn
	e = protocol.SendMsg(conn, &protocol.MsgClientInitiateReconciliation{
		SelfSign: *sign,
	})
	if e != nil {
		fmt.Println(e.Error())
		return // fail
	}
}
