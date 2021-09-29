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
		payobj.address.Address.ToReadable())
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

	// 判断路径选择
	ops := payobj.prequeryMsg.PathForms.PayPaths
	pidx := pathselect - 1
	if pathselect <= 0 || pidx >= len(ops) {
		return "Wrong payment path selected"
	}
	tarpath := ops[pidx]
	//fmt.Println("发起支付", pathselect)
	// 抢占通道支付状态
	if c.user.upstreamSide.ChannelSide.StartBusinessExclusive() == false {
		return "The channel status is occupied. Please try again later"
	}

	// 发送调起支付消息
	randtrsid := rand.Uint64()
	ttimest := time.Now().Unix()
	odidckr := make([]byte, 16)
	rand.Read(odidckr)
	// 总手续费不超过预估手续费的两倍
	maxfee, _ := tarpath.PredictPathFee.Add(&tarpath.PredictPathFee)
	paymsg := &protocol.MsgRequestInitiatePayment{
		TransactionDistinguishId: fields.VarUint8(randtrsid),
		Timestamp:                fields.BlockTxTimestamp(ttimest),
		OrderNoteHashHalfChecker: odidckr,
		HighestAcceptanceFee:     *maxfee,
		PayAmount:                payobj.amount,
		PayeeChannelAddr:         fields.CreateStringMax255(payobj.address.ToReadable(true)),
		TargetPath:               *tarpath.NodeIdPath,
	}

	// 发送消息并创建支付状态机
	se := protocol.SendMsg(c.user.upstreamSide.ChannelSide.WsConn, paymsg)
	if se != nil {
		return "Initiate payment send msg error: " + se.Error()
	}

	// 创建支付状态机
	payaction := chanpay.NewChannelPayActionInstance()
	// 启动日志订阅
	logschan := make(chan *chanpay.PayActionLog, 2)
	go func() {
		for {
			log := <-logschan
			if log == nil || log.IsEnd {
				return // 订阅结束
			}
			// 显示日志
			c.ShowLogString(log.Content, log.IsSuccess, log.IsError)
		}
	}()
	logschan <- &chanpay.PayActionLog{
		IsSuccess: true,
		Content:   fmt.Sprintf("---- start new payment at %s ----", time.Now().Format("2006-01-02 15:04:05")),
	}
	payaction.SubscribeLogs(logschan) // 日志订阅
	// 我发起的支付，我是源头，服务商设置为下游
	payaction.SetDownstreamSide(c.user.upstreamSide)
	// 初始化票据信息
	be := payaction.InitCreateEmptyBillDocumentsByInitPayMsg(paymsg)
	if be != nil {
		payaction.Destroy() // 终止支付，自动解除状态独占
		return "Initiate payment create bill documents error: " + be.Error()
	}
	// 启动消息监听
	isupordown := false // 下游
	payaction.StartOneSideMessageSubscription(isupordown, c.user.upstreamSide.ChannelSide)

	//

	// 暂无错误
	return ""
}

// 取消支付
func (c *ChannelPayClient) BindFuncCancelPayment() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.pendingPaymentObj = nil // 取消
}
