package client

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"strings"
)

// 让前端选择支付
func (c *ChannelPayClient) dealPrequeryPaymentResult(msg *protocol.MsgResponsePrequeryPayment) {
	params := make([]string, 0)
	paths := msg.PathForms.PayPaths
	for _, v := range paths {
		item := fmt.Sprintf(`"%s, total fee: %s"`, strings.Replace(v.Describe.Value(), `"`, ``, -1), v.PredictPathFee.ToFinString())
		params = append(params, item)
	}
	// 支付状态
	c.statusMutex.Lock()
	if c.pendingPaymentObj == nil {
		return // 支付状态已经取消
	}
	c.pendingPaymentObj.prequeryMsg = msg
	payobj := c.pendingPaymentObj
	c.statusMutex.Unlock()

	// 调起选择支付渠道
	checkinfo := fmt.Sprintf(`Transfer %s (%sMei) to %s`,
		payobj.amount.ToFinString(), payobj.amount.ToMeiString(),
		payobj.address.ToReadable())
	c.payui.Eval(fmt.Sprintf(`SelectPaymentPaths("%s", [%s])`,
		checkinfo, strings.Join(params, ",")))
}

// 确定发动交易
// return string 返回错误或者为空
func (c *ChannelPayClient) BindFuncConfirmPayment(pathselect int) string {
	c.statusMutex.Lock()
	payobj := c.pendingPaymentObj
	if payobj == nil {
		return "Pending payment action not find."
	}
	c.statusMutex.Unlock()

	fmt.Println("发起支付", pathselect)
	// TODO: //////

	// 暂无错误
	return ""
}

// 取消支付
func (c *ChannelPayClient) BindFuncCancelPayment() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.pendingPaymentObj = nil // 取消
}
