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

	// 检查状态
	upconn := c.user.upstreamSide.ChannelSide.WsConn
	if c.user.upstreamSide.IsInBusinessExclusive() {
		protocol.SendMsg(upconn, &protocol.MsgBroadcastChannelStatementError{
			ErrCode: 1,
			ErrTip:  fields.CreateStringMax65535("target collection address channel occupied."),
		})
		return
	}

	// 创建支付操作包
	payact := chanpay.NewChannelPayActionInstance()
	payact.SetUpstreamSide(c.user.upstreamSide)

	// 设置我必须签名的地址
	payact.SetMustSignAddresses([]fields.Address{c.user.upstreamSide.ChannelSide.OurAddress})

	// 错误返回
	returnError := func(errmsg string) {
		protocol.SendMsg(upconn, &protocol.MsgBroadcastChannelStatementError{
			ErrCode: 1,
			ErrTip:  fields.CreateStringMax65535(errmsg),
		})
		payact.Destroy() // 资源销毁
	}

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
		Content:   fmt.Sprintf("---- new collecting %s at %s ----", msg.PayAmount.ToFinString(), time.Now().Format("2006-01-02 15:04:05")),
	}
	payact.SubscribeLogs(logschan) // 日志订阅

	// 启动消息监听
	payact.StartOneSideMessageSubscription(true, c.user.upstreamSide.ChannelSide)

	// 初始化票据
	e := payact.InitCreateEmptyBillDocumentsByInitPayMsg(msg)
	if e != nil {
		returnError(e.Error()) // 初始化错误
		return
	}

	// 自动执行签名支付
	return

}
