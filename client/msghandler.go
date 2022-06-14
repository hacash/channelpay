package client

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"strings"
)

/**
 * 消息处理
 */
func (c *ChannelPayClient) startMsgHandler() {
	// Subscription message processing,
	chanobj := make(chan protocol.Message, 1)
	subObj := c.user.servicerStreamSide.ChannelSide.SubscribeMessage(chanobj)
	c.user.msgSubObj = subObj
	// Cycle through messages
	go func() {
		//defer fmt.Println("ChannelPayUser.MsgHandler end")
		for {
			select {
			case v := <-chanobj:
				go c.dealMsg(v)
			case <-subObj.Err():
				c.logoutConnectWindowShow("Network exception. You have logged out") // sign out
				return                                                              // Subscription off
			}
		}
	}()
}

// Message processing
func (c *ChannelPayClient) dealMsg(msg protocol.Message) {
	ty := msg.Type()
	switch ty {
	// Initiate collection
	case protocol.MsgTypeInitiatePayment:
		msgobj := msg.(*protocol.MsgRequestInitiatePayment)
		c.dealInitiatePayment(msgobj)

	// Payment pre query return
	case protocol.MsgTypeResponsePrequeryPayment:
		msgobj := msg.(*protocol.MsgResponsePrequeryPayment)
		if msgobj.ErrCode > 0 {
			// Error display
			c.ShowPaymentErrorString("Prequery payment error: " + msgobj.ErrTip.Value())
			return
		}
		// Call the front-end interface to start payment
		//fmt.Println("PayPathCount: ", msgobj.PathForms.PayPathCount)
		c.dealPrequeryPaymentResult(msgobj)

		// Reconciliation return
	case protocol.MsgTypeServicerRespondReconciliation:
		msgobj := msg.(*protocol.MsgServicerRespondReconciliation)
		c.DealServicerRespondReconciliation(msgobj)

	// Top line
	case protocol.MsgTypeDisplacementOffline:
		c.logoutConnectWindowShow("You have login at another place and this connection has been exited") // sign out
	}

}

// Log out
func (c *ChannelPayClient) logoutConnectWindowShow(tip string) {
	c.user.Logout() // Log out
	// Send an exit login message to the interface
	lgv := c.payui.Eval(fmt.Sprintf(`Logout("%s")`, strings.Replace(tip, `"`, ``, -1))) // 退出
	if DevDebug {
		fmt.Println("Logout() => ", tip, lgv.String())
	}
}
