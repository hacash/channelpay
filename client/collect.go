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

	// 是否关闭收款
	if c.user.servicerStreamSide.ChannelSide.IsInCloseAutoCollectionStatus() {
		returnErrorString("Target account closed collection.")
		return
	}

	// 检查状态
	if c.user.servicerStreamSide.IsInBusinessExclusive() {
		returnErrorString("target collection address channel occupied.")
		return
	}

	// 创建支付操作包
	payact = chanpay.NewChannelPayActionInstance()
	payact.SetUpstreamSide(c.user.servicerStreamSide)
	// 设置签名机
	signmch := NewSignatureMachine(c.user.selfAcc)
	payact.SetSignatureMachine(signmch)
	// 设置我必须签名的地址
	payact.SetMustSignAddresses([]fields.Address{c.user.servicerStreamSide.ChannelSide.OurAddress})

	// 订阅日志，启动日志订阅
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
		Content:   fmt.Sprintf("---- new collecting %s at %s ----", msg.PayAmount.ToFinString(), time.Now().Format("2006-01-02 15:04:05.999")),
	}
	payact.SubscribeLogs(logschan) // 日志订阅

	// 启动消息监听
	payact.StartOneSideMessageSubscription(true, c.user.servicerStreamSide.ChannelSide)

	// 支付、收款成功后的回调
	payact.SetSuccessedBackCall(c.callbackPaymentSuccessed)

	// 初始化票据
	e := payact.InitCreateEmptyBillDocumentsByInitPayMsg(msg)
	if e != nil {
		returnErrorString(e.Error()) // 初始化错误
		return
	}

	// 自动执行签名支付
	return

}
