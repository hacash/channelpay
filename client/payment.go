package client

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"math/rand"
	"strings"
	"time"
)

// Let the front end choose payment
func (c *ChannelPayClient) dealPrequeryPaymentResult(msg *protocol.MsgResponsePrequeryPayment) {
	params := make([]string, 0)
	paths := msg.PathForms.PayPaths
	for _, v := range paths {
		feeshow := fmt.Sprintf(`%s`, v.PredictPathFeeAmt.ToFinString())
		feesat := v.PredictPathFeeSat.GetRealSatoshi()
		if feesat > 0 {
			feeshow = fmt.Sprintf(`%d sats`, feesat)
		}
		item := fmt.Sprintf(`"%s, total fee: %s"`, strings.Replace(v.Describe.Value(), `"`, ``, -1), feeshow)
		params = append(params, item)
	}
	// Payment status
	c.statusMutex.Lock()
	if c.pendingPaymentObj == nil {
		return // Payment status cancelled
	}
	c.pendingPaymentObj.prequeryMsg = msg
	payobj := c.pendingPaymentObj
	c.statusMutex.Unlock()

	// Select payment channel
	payshow := []string{}
	if payobj.amount.IsPositive() {
		payshow = append(payshow, fmt.Sprintf(`%s (%sMei)`,
			payobj.amount.ToFinString(),
			payobj.amount.ToMeiString()))
	}
	// if SAT
	if payobj.satoshi.GetRealSatoshi() > 0 {
		payshow = append(payshow, fmt.Sprintf(`%d sats`,
			payobj.satoshi.GetRealSatoshi()))
	}
	checkinfo := fmt.Sprintf(`Transfer %s to %s`,
		strings.Join(payshow, " and "),
		payobj.address.Address.ToReadable())

	c.payui.Eval(fmt.Sprintf(`SelectPaymentPaths("%s", [%s])`,
		checkinfo, strings.Join(params, ",")))
}

// Confirm to launch transaction
// Return string returns an error or is empty
func (c *ChannelPayClient) BindFuncConfirmPayment(pathselect int) string {
	c.statusMutex.Lock()
	payobj := c.pendingPaymentObj
	if payobj == nil {
		return "Pending payment action not find."
	}
	c.statusMutex.Unlock()

	// Judge path selection
	ops := payobj.prequeryMsg.PathForms.PayPaths
	pidx := pathselect - 1
	if pathselect <= 0 || pidx >= len(ops) {
		return "Wrong payment path selected"
	}
	tarpath := ops[pidx]
	//fmt.Println("发起支付", pathselect)
	// Preemption channel payment status
	if c.user.servicerStreamSide.ChannelSide.StartBusinessExclusive() == false {
		return "The channel status is occupied. Please try again later"
	}

	// Send call up payment message
	randtrsid := rand.Uint64()
	ttimest := time.Now().Unix()
	odidckr := make([]byte, 16)
	rand.Read(odidckr)
	// The total handling fee shall not exceed twice the estimated handling fee
	maxfee, _ := tarpath.PredictPathFeeAmt.Add(&tarpath.PredictPathFeeAmt)
	paymsg := &protocol.MsgRequestInitiatePayment{
		TransactionDistinguishId: fields.VarUint8(randtrsid),
		Timestamp:                fields.BlockTxTimestamp(ttimest),
		OrderNoteHashHalfChecker: odidckr,
		HighestAcceptanceFee:     *maxfee,
		PayAmount:                payobj.amount,
		PaySatoshi:               payobj.satoshi,
		PayeeChannelAddr:         fields.CreateStringMax255(payobj.address.ToReadable(true)),
		TargetPath:               *tarpath.NodeIdPath,
	}

	// Create payment state machine
	payaction := chanpay.NewChannelPayActionInstance()
	// Start log subscription
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
		Content:   fmt.Sprintf("---- start new payment at %s ----", time.Now().Format("2006-01-02 15:04:05.999")),
	}
	payaction.SubscribeLogs(logschan) // Log subscription

	// I am the source of payment initiated by me, and the service provider is set as downstream
	payaction.SetDownstreamSide(c.user.servicerStreamSide)

	// Setting up a signer
	signmch := NewSignatureMachine(c.user.selfAcc)
	payaction.SetSignatureMachine(signmch)
	// Set the address I must sign
	payaction.SetMustSignAddresses([]fields.Address{c.user.servicerStreamSide.ChannelSide.OurAddress})

	// Start message listening
	isupordown := false // 下游
	payaction.StartOneSideMessageSubscription(isupordown, c.user.servicerStreamSide.ChannelSide)

	// Callback after successful payment and collection
	payaction.SetSuccessedBackCall(c.callbackPaymentSuccessed)

	// Initialize ticket information
	be := payaction.InitCreateEmptyBillDocumentsByInitPayMsg(paymsg)
	if be != nil {
		payaction.Destroy() // Terminate the payment and automatically remove the exclusive status
		return "Initiate payment create bill documents error: " + be.Error()
	}

	// Send message and create payment state machine
	se := protocol.SendMsg(c.user.servicerStreamSide.ChannelSide.WsConn, paymsg)
	if se != nil {
		payaction.Destroy() // Terminate the payment and automatically remove the exclusive status
		return "Initiate payment send msg error: " + se.Error()
	}

	// Waiting for payment response

	// No error
	return ""
}

// Cancel payment
func (c *ChannelPayClient) BindFuncCancelPayment() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.pendingPaymentObj = nil // cancel
}
