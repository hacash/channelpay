package client

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"time"
)

/**
 * 处理收款
 */
func (c *ChannelPayClient) dealInitiatePayment(msg *protocol.MsgRequestInitiatePayment) {

	//c.ShowLogString(fmt.Sprintf("collecting %s ...", msg.PayAmount.ToFinString()), false, false)

	var payact *chanpay.ChannelPayActionInstance = nil
	returnErrorString := func(err string) {
		upconn := c.user.servicerStreamSide.ChannelSide.WsConn
		protocol.SendMsg(upconn, &protocol.MsgBroadcastChannelStatementError{
			ErrCode: 1,
			ErrTip:  fields.CreateStringMax65535(err),
		})
		if payact != nil {
			payact.Destroy()
		}
	}

	// Close collection
	if c.user.servicerStreamSide.ChannelSide.IsInCloseAutoCollectionStatus() {
		returnErrorString("Target account closed collection.")
		return
	}

	// Check status
	if c.user.servicerStreamSide.IsInBusinessExclusive() {
		returnErrorString("target collection address channel occupied.")
		return
	}

	// Create payment operation package
	payact = chanpay.NewChannelPayActionInstance()
	payact.SetUpstreamSide(c.user.servicerStreamSide)
	// Setting up a signer
	signmch := NewSignatureMachine(c.user.selfAcc)
	payact.SetSignatureMachine(signmch)
	// Set the address I must sign
	payact.SetMustSignAddresses([]fields.Address{c.user.servicerStreamSide.ChannelSide.OurAddress})

	// Log subscription, start log subscription
	logschan := make(chan *chanpay.PayActionLog, 2)
	go func() {
		for {
			log := <-logschan
			if log == nil || log.IsEnd {
				return // End of subscription
			}
			// Show log
			c.ShowLogString(log.Content, log.IsSuccess, log.IsError)
		}
	}()
	logschan <- &chanpay.PayActionLog{
		IsSuccess: true,
		Content:   fmt.Sprintf("---- new collecting %s at %s ----", msg.PayAmount.ToFinString(), time.Now().Format("2006-01-02 15:04:05.999")),
	}
	payact.SubscribeLogs(logschan) // Log subscription

	// Start message listening
	payact.StartOneSideMessageSubscription(true, c.user.servicerStreamSide.ChannelSide)

	// Callback after successful payment and collection
	payact.SetSuccessedBackCall(c.callbackPaymentSuccessed)

	// Initialize ticket
	e := payact.InitCreateEmptyBillDocumentsByInitPayMsg(msg)
	if e != nil {
		returnErrorString(e.Error()) // Initialization error
		return
	}

	// Automatic signature payment
	return

}
