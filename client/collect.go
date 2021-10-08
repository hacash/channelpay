package client

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
)

/**
 * 处理收款
 */
func (c *ChannelPayClient) dealInitiatePayment(msg *protocol.MsgRequestInitiatePayment) {

	c.ShowLogString(fmt.Sprintf("got collect msg, amt: %s", msg.PayAmount.ToFinString()), false, false)

	// 检查状态
	upconn := c.user.upstreamSide.ChannelSide.WsConn
	if c.user.upstreamSide.IsInBusinessExclusive() {
		protocol.SendMsg(upconn, &protocol.MsgBroadcastChannelStatementError{
			ErrCode: 1,
			ErrTip:  fields.CreateStringMax65535("target collection address channel occupied."),
		})
		return
	}

	protocol.SendMsg(upconn, &protocol.MsgBroadcastChannelStatementError{
		ErrCode: 1,
		ErrTip:  fields.CreateStringMax65535("okokok!!!."),
	})

}
